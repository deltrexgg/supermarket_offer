// Line-art Gulf skyline for the footer: towers, a sail-shaped hotel, a mosque
// dome with minaret and a dhow on the water. Decorative only.
export default function Skyline() {
  return (
    <svg className="sky" viewBox="0 0 1200 120" preserveAspectRatio="xMidYMax slice" aria-hidden="true">
      <g fill="none" stroke="#188540" strokeWidth="2" strokeLinejoin="round" strokeLinecap="round">
        <path d="M0 112H1200" />
        <path d="M30 112V78H70V112M44 78V66H56V78" />
        <path d="M90 112V60H118V112M96 70H112M96 82H112M96 94H112" />
        <path d="M150 112V86H200V112" />
        <path d="M250 112V84H330V112M258 84Q290 42 322 84M290 63V54" />
        <path d="M346 112V58H356V112M342 58H360M347 58V46H355V58M351 46V36" />
        <path d="M390 112V70H430V112M400 70V62H420V70" />
        <path d="M460 112V88H520V112" />
        <path d="M560 112L568 60H576L580 30H586L590 4L594 30H600L604 60H612L620 112" />
        <path d="M640 112V74H680V112" />
        <path d="M700 112V52H730V112M706 62H724M706 76H724M706 90H724" />
        <path d="M780 112V24M780 24Q842 52 832 112M780 52H822" />
        <path d="M860 112V80H910V112" />
        <path d="M975 100H1065L1052 110H988ZM1015 100V48M1015 54L1050 96H1015" />
        <path d="M1100 112V66H1140V112M1150 112V90H1190V112" />
      </g>
    </svg>
  );
}
