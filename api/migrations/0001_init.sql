-- 0001_init: core schema for the supermarket offers platform.
-- Runs inside a single transaction (see internal/migrate), so it is all-or-nothing.

CREATE EXTENSION IF NOT EXISTS pg_trgm;   -- fast "milk" / "almarai" search
CREATE EXTENSION IF NOT EXISTS citext;    -- case-insensitive emails and names

CREATE TYPE user_role          AS ENUM ('admin', 'supermarket');
CREATE TYPE supermarket_status AS ENUM ('pending', 'approved', 'blocked');

CREATE TABLE areas (
  id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name citext NOT NULL UNIQUE
);

CREATE TABLE categories (
  id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name citext NOT NULL UNIQUE
);

CREATE TABLE supermarkets (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name       varchar(150) NOT NULL,
  slug       varchar(150) NOT NULL UNIQUE,
  logo_path  text,
  phone      varchar(30),
  status     supermarket_status NOT NULL DEFAULT 'pending',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Passwords live in Supabase Auth. This table only maps a Supabase user to a role
-- and a supermarket. auth_user_id is the Supabase user's id (no foreign key, so the
-- migration also applies to a plain local Postgres).
CREATE TABLE users (
  id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  auth_user_id   uuid   NOT NULL UNIQUE,
  email          citext NOT NULL UNIQUE,
  role           user_role NOT NULL,
  supermarket_id bigint UNIQUE REFERENCES supermarkets(id) ON DELETE CASCADE,
  created_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT users_role_tenant CHECK (
    (role = 'admin'       AND supermarket_id IS NULL) OR
    (role = 'supermarket' AND supermarket_id IS NOT NULL)
  )
);

CREATE TABLE branches (
  id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  supermarket_id bigint NOT NULL REFERENCES supermarkets(id) ON DELETE CASCADE,
  area_id        bigint NOT NULL REFERENCES areas(id) ON DELETE RESTRICT,
  name           varchar(150) NOT NULL,
  address        text NOT NULL,
  phone          varchar(30),
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (supermarket_id, name),
  UNIQUE (id, supermarket_id)                     -- target for the tenant-safe FK below
);
CREATE INDEX branches_area_idx ON branches (area_id);

CREATE TABLE offers (
  id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  supermarket_id bigint NOT NULL REFERENCES supermarkets(id) ON DELETE CASCADE,
  category_id    bigint NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  item_name      varchar(200) NOT NULL,
  item_name_ar   varchar(200),                    -- AI-translated (nice-to-have)
  brand          varchar(100),
  unit           varchar(50) NOT NULL,            -- "1 L", "5 kg"
  description    text,                            -- AI-written (nice-to-have)
  description_ar text,
  regular_price  numeric(10,2) NOT NULL CHECK (regular_price > 0),   -- AED
  offer_price    numeric(10,2) NOT NULL CHECK (offer_price  > 0),
  discount_pct   smallint GENERATED ALWAYS AS
                   (round((regular_price - offer_price) * 100 / regular_price)::smallint) STORED,
  valid_from     date NOT NULL,                   -- Dubai calendar days, inclusive
  valid_to       date NOT NULL,
  photo_path     text,
  poster_path    text,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT offer_price_lt_regular CHECK (offer_price < regular_price),
  CONSTRAINT offer_dates_ordered    CHECK (valid_to >= valid_from),
  UNIQUE (id, supermarket_id)
);
CREATE INDEX offers_validity_idx ON offers (valid_from, valid_to);
CREATE INDEX offers_tenant_idx   ON offers (supermarket_id, valid_to DESC);
CREATE INDEX offers_category_idx ON offers (category_id);
CREATE INDEX offers_item_trgm    ON offers USING gin (item_name gin_trgm_ops);
CREATE INDEX offers_brand_trgm   ON offers USING gin (brand     gin_trgm_ops);

-- Which branches an offer applies to. The composite FKs make it impossible
-- to link an offer to a branch that belongs to a different supermarket.
CREATE TABLE offer_branches (
  offer_id       bigint NOT NULL,
  branch_id      bigint NOT NULL,
  supermarket_id bigint NOT NULL,
  PRIMARY KEY (offer_id, branch_id),
  FOREIGN KEY (offer_id,  supermarket_id) REFERENCES offers   (id, supermarket_id) ON DELETE CASCADE,
  FOREIGN KEY (branch_id, supermarket_id) REFERENCES branches (id, supermarket_id) ON DELETE CASCADE
);
CREATE INDEX offer_branches_branch_idx ON offer_branches (branch_id);

-- Status is derived, never stored, so it cannot go stale. "Today" is Dubai time.
CREATE VIEW offers_with_status AS
SELECT o.*,
       CASE WHEN d.today < o.valid_from THEN 'upcoming'
            WHEN d.today > o.valid_to   THEN 'expired'
            ELSE 'active' END AS status,
       (o.valid_to - d.today) AS days_left        -- 0 means "Ends today"
FROM offers o
CROSS JOIN LATERAL (SELECT (now() AT TIME ZONE 'Asia/Dubai')::date AS today) d;

CREATE FUNCTION set_updated_at() RETURNS trigger LANGUAGE plpgsql SET search_path = '' AS
$$ BEGIN NEW.updated_at = now(); RETURN NEW; END $$;

CREATE TRIGGER supermarkets_touch BEFORE UPDATE ON supermarkets
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER branches_touch BEFORE UPDATE ON branches
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER offers_touch BEFORE UPDATE ON offers
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Supabase exposes public tables to the internet through its Data API. Row-level
-- security with no policies denies every Data API request. The Go API connects as the
-- postgres role, which bypasses these rules, so the app is unaffected.
ALTER TABLE areas          ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories     ENABLE ROW LEVEL SECURITY;
ALTER TABLE supermarkets   ENABLE ROW LEVEL SECURITY;
ALTER TABLE users          ENABLE ROW LEVEL SECURITY;
ALTER TABLE branches       ENABLE ROW LEVEL SECURITY;
ALTER TABLE offers         ENABLE ROW LEVEL SECURITY;
ALTER TABLE offer_branches ENABLE ROW LEVEL SECURITY;

-- A view runs with its owner's rights by default, which would skip the rules above.
-- security_invoker makes it obey the caller's rights instead.
ALTER VIEW offers_with_status SET (security_invoker = true);
