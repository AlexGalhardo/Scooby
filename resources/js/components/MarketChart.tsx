import { useEffect, useRef } from "react";
import { createChart, CandlestickSeries, LineSeries, ColorType, type UTCTimestamp } from "lightweight-charts";
import type { Candle, Point } from "@/types";
export default function MarketChart({candles,curve}:{candles?:Candle[];curve?:Point[]}){
 const ref=useRef<HTMLDivElement>(null);
 useEffect(()=>{if(!ref.current)return;const dark=document.documentElement.classList.contains("dark");const chart=createChart(ref.current,{height:300,layout:{background:{type:ColorType.Solid,color:"transparent"},textColor:dark?"#9ca3af":"#6b7280",attributionLogo:true},grid:{vertLines:{color:dark?"#ffffff06":"#00000006"},horzLines:{color:dark?"#ffffff08":"#00000008"}},rightPriceScale:{borderVisible:false},timeScale:{borderVisible:false,timeVisible:true},handleScroll:true});
 const timestamp=(time:string)=>Math.floor(new Date(time).getTime()/1000) as UTCTimestamp;
 if(candles){chart.addSeries(CandlestickSeries,{upColor:"#34d399",downColor:"#f87171",borderVisible:false,wickUpColor:"#34d399",wickDownColor:"#f87171"}).setData(candles.map(c=>({time:timestamp(c.time),open:Number(c.open),high:Number(c.high),low:Number(c.low),close:Number(c.close)})))}
 if(curve){chart.addSeries(LineSeries,{color:"#34d399",lineWidth:2,title:"Equity"}).setData(curve.map(p=>({time:timestamp(p.time),value:Number(p.equity)})));chart.addSeries(LineSeries,{color:"#64748b",lineWidth:1,title:"Benchmark"}).setData(curve.map(p=>({time:timestamp(p.time),value:Number(p.benchmark)})))}
 chart.timeScale().fitContent();const observer=new ResizeObserver(entries=>chart.applyOptions({width:entries[0].contentRect.width}));observer.observe(ref.current);return()=>{observer.disconnect();chart.remove()}
 },[candles,curve]);return <div ref={ref} className="w-full" aria-label="Gráfico de mercado; valores detalhados nas tabelas"/>
}
