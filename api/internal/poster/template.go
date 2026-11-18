package poster

import "html/template"

var page = template.Must(template.New("poster").Parse(pageHTML))

const pageHTML = `<!DOCTYPE html>
<html lang="{{if .RTL}}ar{{else}}en{{end}}" dir="{{if .RTL}}rtl{{else}}ltr{{end}}">
<head>
<meta charset="utf-8">
<title>Offer poster</title>
<style>
@page{size:A4;margin:0}
*{box-sizing:border-box}
html,body{margin:0;padding:0;background:#fff}
body{font-family:"Noto Sans","DejaVu Sans",sans-serif;-webkit-print-color-adjust:exact;print-color-adjust:exact}
.rtl,.rtl *{font-family:"Noto Kufi Arabic","Noto Naskh Arabic","Noto Sans Arabic","Noto Sans",sans-serif}
.frame{container-type:inline-size;{{if .Print}}width:210mm;height:297mm;{{else}}width:1240px;height:1754px;{{end}}overflow:hidden}
.p{--u:1cqw;width:100%;height:100%;display:flex;flex-direction:column;overflow:hidden;position:relative}
.ltr{direction:ltr;unicode-bidi:isolate}
.clamp{display:-webkit-box;-webkit-line-clamp:3;-webkit-box-orient:vertical;overflow:hidden}
.ph,.lg{display:grid;place-items:center;border-radius:50%;font-weight:800}
.photo{flex:1;min-height:0;display:grid;place-items:center;position:relative}
.photo img{max-width:70%;max-height:92%;object-fit:contain}
.badge{position:absolute;display:grid;place-content:center;text-align:center;border-radius:50%;font-weight:800;line-height:1}
.pnum{font-weight:800;letter-spacing:-.03em;line-height:.9}
.hot{background:linear-gradient(165deg,#FF7A30 0%,#E8481C 52%,#C42B17 100%);color:#fff}
.hot .head{display:flex;align-items:center;gap:calc(var(--u)*2.4);background:#fff;color:#14301F;padding:calc(var(--u)*2.4) calc(var(--u)*4.5)}
.hot .lg{width:calc(var(--u)*8.5);height:calc(var(--u)*8.5);background:#0E7A4B;color:#fff;font-size:calc(var(--u)*4.2);flex:none}
.hot .sm{font-weight:800;font-size:calc(var(--u)*3.4);line-height:1.2}
.hot .sub{font-size:calc(var(--u)*2.1);opacity:.7;margin-top:calc(var(--u)*.4)}
.hot .title{text-align:center;padding-top:calc(var(--u)*3.5);font-weight:800;font-size:calc(var(--u)*14);text-transform:uppercase;transform:rotate(-3deg);text-shadow:0 calc(var(--u)*.8) 0 rgba(120,20,0,.35)}
.rtl.hot .title,.rtl .hot .title{text-transform:none;font-size:calc(var(--u)*12);line-height:1.3}
.hot .ph{width:calc(var(--u)*30);height:calc(var(--u)*30);background:rgba(255,255,255,.9);color:#C42B17;font-size:calc(var(--u)*15)}
.hot .badge{top:calc(var(--u)*1);inset-inline-end:calc(var(--u)*7);width:calc(var(--u)*25);height:calc(var(--u)*25);background:#FFD23F;color:#7A1200;transform:rotate(8deg);box-shadow:0 calc(var(--u)*1) calc(var(--u)*2) rgba(90,15,0,.35)}
.hot .badge b{font-size:calc(var(--u)*9)}
.hot .badge small{font-size:calc(var(--u)*3.4);font-weight:800;letter-spacing:.04em}
.info{text-align:center;padding:0 calc(var(--u)*5)}
.brand{font-size:calc(var(--u)*2.7);letter-spacing:.22em;text-transform:uppercase;font-weight:700}
.rtl .brand{letter-spacing:0}
.unit,.brand,.item{unicode-bidi:plaintext}
.item{font-weight:800;font-size:calc(var(--u)*6.6);line-height:1.15;margin:calc(var(--u)*.8) 0}
.unit{font-size:calc(var(--u)*3.4);font-weight:600}
.hot .price{display:flex;align-items:baseline;justify-content:center;gap:calc(var(--u)*3);padding:calc(var(--u)*2) 0 calc(var(--u)*1)}
.was{text-decoration:line-through;text-decoration-thickness:.12em;font-weight:600;font-size:calc(var(--u)*4.2)}
.now{display:flex;align-items:baseline;gap:calc(var(--u)*1)}
.cur{font-weight:800;font-size:calc(var(--u)*5)}
.hot .big{font-size:calc(var(--u)*22)}
.hot .dec{font-size:calc(var(--u)*9)}
.hot .valid{align-self:center;background:rgba(0,0,0,.28);border-radius:999px;padding:calc(var(--u)*1.1) calc(var(--u)*4);font-size:calc(var(--u)*3);font-weight:700;margin:calc(var(--u)*1) 0 calc(var(--u)*3)}
.hot .banner{align-self:center;background:#FFD23F;color:#7A1200;border-radius:999px;padding:calc(var(--u)*.8) calc(var(--u)*3.5);font-size:calc(var(--u)*3);font-weight:800;margin-top:calc(var(--u)*-1);margin-bottom:calc(var(--u)*1.5)}
.hot .foot{background:#14301F;color:#E8F3EC;padding:calc(var(--u)*2.4) calc(var(--u)*4.5);font-size:calc(var(--u)*2.4);line-height:1.5}
.hot .foot b{color:#fff}
.wk{background:#10243A;color:#fff}
.wk .top{padding:calc(var(--u)*5) calc(var(--u)*5) 0}
.wk .tag{display:inline-block;background:#FFC93C;color:#10243A;font-weight:800;font-size:calc(var(--u)*2.4);letter-spacing:.18em;text-transform:uppercase;padding:calc(var(--u)*.9) calc(var(--u)*2.4);border-radius:calc(var(--u)*.8)}
.rtl .wk .tag{letter-spacing:0}
.wk h1{margin:calc(var(--u)*2) 0 0;font-weight:800;font-size:calc(var(--u)*8);line-height:1.05;color:#FFC93C;text-transform:uppercase}
.rtl .wk h1{text-transform:none;font-size:calc(var(--u)*8);line-height:1.3}
.wk .when{font-size:calc(var(--u)*3);font-weight:600;margin-top:calc(var(--u)*1.5);opacity:.9}
.wk .stage:before{content:"";position:absolute;width:calc(var(--u)*46);height:calc(var(--u)*46);border-radius:50%;background:#1E4468;border:calc(var(--u)*1.2) dashed #FFC93C}
.wk .photo img{position:relative}
.wk .ph{position:relative;width:calc(var(--u)*34);height:calc(var(--u)*34);background:#FFC93C;color:#10243A;font-size:calc(var(--u)*16)}
.wk .badge{top:calc(var(--u)*2);inset-inline-end:calc(var(--u)*8);width:calc(var(--u)*20);height:calc(var(--u)*20);background:#FF5A3C;color:#fff;transform:rotate(-8deg)}
.wk .badge b{font-size:calc(var(--u)*7)}
.wk .badge small{font-size:calc(var(--u)*3);font-weight:800}
.wk .sheet{background:#fff;color:#10243A;border-radius:calc(var(--u)*6) calc(var(--u)*6) 0 0;padding:calc(var(--u)*5) calc(var(--u)*5) 0}
.wk .row{display:flex;justify-content:space-between;align-items:flex-end;gap:calc(var(--u)*3)}
.wk .brand{color:#0E7A4B}
.wk .item{font-size:calc(var(--u)*5.6);margin:calc(var(--u)*.6) 0 calc(var(--u)*.4)}
.wk .unit{color:#52657A}
.wk .txt{text-align:start;min-width:0}
.wk .pr{text-align:end;flex:none}
.wk .was{font-size:calc(var(--u)*3.4);color:#8A98A8;display:block}
.wk .now{justify-content:flex-end;color:#0E7A4B}
.wk .cur{font-size:calc(var(--u)*4.2)}
.wk .big{font-size:calc(var(--u)*17)}
.wk .dec{font-size:calc(var(--u)*7)}
.wk .meta{margin-top:calc(var(--u)*3);border-top:calc(var(--u)*.35) solid #E3E9EF;padding:calc(var(--u)*2.2) 0 calc(var(--u)*3);display:flex;justify-content:space-between;gap:calc(var(--u)*3);font-size:calc(var(--u)*2.4);color:#40546A;line-height:1.5}
.wk .meta b{color:#10243A}
.wk .band{background:#0E7A4B;color:#fff;display:flex;align-items:center;gap:calc(var(--u)*2.4);padding:calc(var(--u)*2.4) calc(var(--u)*5)}
.wk .band .lg{width:calc(var(--u)*7);height:calc(var(--u)*7);background:#fff;color:#0E7A4B;font-size:calc(var(--u)*3.4);flex:none}
.wk .band .sm{font-weight:800;font-size:calc(var(--u)*3.2)}
.wk .band .sub{font-size:calc(var(--u)*2.1);opacity:.9}
.wk .band .spacer{flex:1}
.wk .banner{background:#FFC93C;color:#10243A;border-radius:999px;padding:calc(var(--u)*.7) calc(var(--u)*3);font-size:calc(var(--u)*2.6);font-weight:800}
</style>
</head>
<body class="{{if .RTL}}rtl{{end}}">
<div class="frame"><div class="p {{if eq .Template "hot_deal"}}hot{{else}}wk{{end}}">
{{if eq .Template "hot_deal"}}
<div class="head"><div class="lg">{{.Initial}}</div><div><div class="sm">{{.Shop}}</div><div class="sub">{{.Areas}}</div></div></div>
<div class="title">{{.L.HotDeal}}</div>
<div class="photo">{{if .PhotoURI}}<img src="{{.PhotoURI}}" alt="">{{else}}<div class="ph">{{.ItemInitial}}</div>{{end}}
{{if gt .Discount 0}}<div class="badge"><b class="ltr">{{.Discount}}%</b><small>{{.L.Off}}</small></div>{{end}}</div>
<div class="info">{{if .Brand}}<div class="brand">{{.Brand}}</div>{{end}}<div class="item clamp">{{.Item}}</div><div class="unit">{{.Unit}}</div></div>
<div class="price"><span class="was ltr">{{.L.Currency}} {{.Was}}</span><span class="now ltr"><span class="cur">{{.L.Currency}}</span><span class="pnum big">{{.PriceInt}}</span><span class="pnum dec">{{.PriceDec}}</span></span></div>
{{if .Banner}}<div class="banner">{{.Banner}}</div>{{end}}
<div class="valid">{{.Valid}}</div>
<div class="foot"><b>{{.L.Available}}:</b> {{.Branches}}</div>
{{else}}
<div class="top"><span class="tag">{{.L.Weekend}}</span><h1>{{.Shop}}</h1><div class="when">{{.Valid}}</div></div>
<div class="photo stage">{{if .PhotoURI}}<img src="{{.PhotoURI}}" alt="">{{else}}<div class="ph">{{.ItemInitial}}</div>{{end}}
{{if gt .Discount 0}}<div class="badge"><b class="ltr">{{.Discount}}%</b><small>{{.L.Off}}</small></div>{{end}}</div>
<div class="sheet">
<div class="row"><div class="txt">{{if .Brand}}<div class="brand">{{.Brand}}</div>{{end}}<div class="item clamp">{{.Item}}</div><div class="unit">{{.Unit}}</div></div>
<div class="pr"><span class="was ltr">{{.L.Currency}} {{.Was}}</span><span class="now ltr"><span class="cur">{{.L.Currency}}</span><span class="pnum big">{{.PriceInt}}</span><span class="pnum dec">{{.PriceDec}}</span></span></div></div>
<div class="meta"><span><b>{{.L.Available}}:</b> {{.Branches}}</span></div>
</div>
<div class="band"><div class="lg">{{.Initial}}</div><div><div class="sm">{{.Shop}}</div><div class="sub">{{.Areas}}</div></div><span class="spacer"></span>{{if .Banner}}<span class="banner">{{.Banner}}</span>{{end}}</div>
{{end}}
</div></div>
{{if .Print}}<script>window.addEventListener("load",function(){setTimeout(function(){window.print()},300)})</script>{{end}}
</body>
</html>`
