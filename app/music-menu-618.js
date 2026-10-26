/* Fit labels to the available column, leaving the notification count visible. */
(()=>{
 'use strict';
 const previousItems=navItems;navItems=function(user){const items=previousItems.call(this,user);if(!['admin','coordinador_sgm'].includes(user?.role))return items;return [...items.filter(item=>item[0]!=='settings'),...items.filter(item=>item[0]==='settings')];};window.navItems=navItems;
 let frame=0;
 function fit(){
  cancelAnimationFrame(frame);frame=requestAnimationFrame(()=>{
   for(const span of document.querySelectorAll('.sidebar #nav button[data-page]>span')){
    const base=document.documentElement.classList.contains('music615-large-text')?14:13;span.style.removeProperty('font-size');let size=base;
    while(span.scrollWidth>span.clientWidth+1&&size>10){size-=.25;span.style.setProperty('font-size',size+'px','important')}
   }
  });
 }
 const nav=renderNav;renderNav=function(...args){const value=nav.apply(this,args);fit();return value};window.renderNav=renderNav;
 window.addEventListener('resize',fit);document.fonts.ready.then(fit);
 const observer=new ResizeObserver(fit);const start=()=>{const host=document.querySelector('.sidebar');if(host)observer.observe(host);fit()};
 if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',start,{once:true});else start();
})();
