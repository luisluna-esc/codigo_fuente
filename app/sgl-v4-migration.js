(function(){
'use strict';
const pairs=[
 ['Mi Clase Guía Mayor Máster','Guía Mayor Máster'],['Mi Guía Mayor Máster','Guía Mayor Máster'],
 ['Mi Clase Guía Mayor','Guía Mayor'],['Mi Guía Mayor','Guía Mayor'],['Mi Clase Agrupada','Clase Agrupada'],
 ['Clase Agrupada Oficial','Sistema de Gestión de Música']
];
function fix(root){
 const w=document.createTreeWalker(root||document.body,NodeFilter.SHOW_TEXT);let n;
 while((n=w.nextNode())){if(n.parentElement?.closest('#req_text,[contenteditable="true"]'))continue;let s=n.nodeValue;for(const [a,b] of pairs)s=s.split(a).join(b);if(s!==n.nodeValue)n.nodeValue=s;}
 const images=[...(root?.matches?.('img')?[root]:[]),...(root||document.body).querySelectorAll('img')];images.forEach(img=>{if(img.closest('#req_text'))return;const src=img.getAttribute('src')||'',key=((img.id||'')+' '+(img.className||'')+' '+(img.alt||'')+' '+(src.startsWith('data:')?'':src)).toLowerCase();if(/logo|escudo|emblem|conquist/.test(key)){img.style.objectFit='contain';img.style.width='auto';img.style.height='auto';img.style.maxWidth='100%';}});
}
function migrateStorage(){
 const swaps=[['ClaseAgrupada','MinisterioMusica'],['claseagrupada','sistemagestionliderazgo'],['clase-agrupada','sistema-gestion-liderazgo']];
 const keys=[];for(let i=0;i<localStorage.length;i++)keys.push(localStorage.key(i));
 for(const key of keys){for(const [a,b] of swaps){if(key&&key.includes(a)){const nk=key.split(a).join(b);if(localStorage.getItem(nk)==null)localStorage.setItem(nk,localStorage.getItem(key));}}}
}
const boot=()=>{migrateStorage();fix(document.body);new MutationObserver(m=>{const roots=new Set(m.flatMap(x=>[...x.addedNodes]).filter(n=>n.nodeType===1&&n.isConnected&&!n.closest('#req_text')));for(const n of roots){let p=n.parentElement,covered=false;while(p){if(roots.has(p)){covered=true;break;}p=p.parentElement;}if(!covered)fix(n);}}).observe(document.body,{childList:true,subtree:true});};if(document.readyState==='loading')window.addEventListener('DOMContentLoaded',boot,{once:true});else boot();
})();