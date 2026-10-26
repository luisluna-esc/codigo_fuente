/* Multiple object selection stays in the editor and never changes the template schema. */
window.MusicSelection621=({model,primary,setPrimary,history,redraw,controls,edit})=>{
 let items=[],clipboard=[];
 const key=r=>r.kind+':'+r.index,ordinary=r=>['field','image'].includes(r.kind);
 const valid=r=>ordinary(r)?!!(r.kind==='field'?model().fields:model().images610)[r.index]:['qr','verification','background'].includes(r.kind);
 const data=r=>r.kind==='field'?model().fields[r.index]:r.kind==='image'?model().images610[r.index]:r.kind==='qr'?{x:model().qrX,y:model().qrY}:model().verification618;
 function reset(){items=[{...primary()}]}
 function sync(){items=items.filter(valid);if(!items.some(r=>key(r)===key(primary())))reset()}
 function snapshot(){sync();return items.map(r=>({...r}))}
 function restore(value){items=(value||[{...primary()}]).map(r=>({...r}));sync()}
 function has(kind,index){return items.some(r=>r.kind===kind&&r.index===index)}
 function paint(stage){sync();stage.dataset.multiple=String(items.length>1);for(const box of stage.querySelectorAll('[data-kind]'))box.classList.toggle('selected',has(box.dataset.kind,Number(box.dataset.item)))}
 function pick(ref,add=false){
  sync();const exists=items.some(r=>key(r)===key(ref));
  if(add&&ordinary(ref)){
   items=items.filter(ordinary);items=exists?items.filter(r=>key(r)!==key(ref)):[...items,ref];
   setPrimary(items.at(-1)||{kind:'background',index:0});
  }else{if(!exists||!ordinary(ref))items=[ref];setPrimary(ref)}
 }
 function bounds(stage){
  const r=stage.getBoundingClientRect();let minX=-Infinity,maxX=Infinity,minY=-Infinity,maxY=Infinity;
  for(const ref of items){const box=stage.querySelector(`[data-kind="${ref.kind}"][data-item="${ref.index}"]`);if(!box)continue;const b=box.getBoundingClientRect();minX=Math.max(minX,(r.left-b.left)/r.width*100);maxX=Math.min(maxX,(r.right-b.right)/r.width*100);minY=Math.max(minY,(r.top-b.top)/r.height*100);maxY=Math.min(maxY,(r.bottom-b.bottom)/r.height*100)}
  return {minX,maxX,minY,maxY};
 }
 function translate(stage,original,limit,dx,dy){if(window.SGL592?.locked622())return;
  dx=Math.max(limit.minX,Math.min(limit.maxX,dx));dy=Math.max(limit.minY,Math.min(limit.maxY,dy));
  for(const {ref,x,y}of original){const f=data(ref);if(ref.kind==='qr'){model().qrX=x+dx;model().qrY=y+dy}else{f.x=x+dx;f.y=y+dy}const node=stage.querySelector(`[data-kind="${ref.kind}"][data-item="${ref.index}"]`);if(node){node.style.left=(x+dx)+'%';node.style.top=(y+dy)+'%'}}
 }
 function positions(){return items.filter(r=>r.kind!=='background').map(ref=>({ref:{...ref},x:data(ref).x,y:data(ref).y}))}
 function focus(stage){(stage.querySelector(`[data-kind="${primary().kind}"][data-item="${primary().index}"]`)||stage).focus({preventScroll:true})}
 function bind(stage,paper,before){
  paint(stage);
  for(const box of stage.querySelectorAll('[data-kind]')){
   const ref={kind:box.dataset.kind,index:Number(box.dataset.item)};
   box.ondblclick=event=>{if(ref.kind!=='field'||event.ctrlKey||event.metaKey||event.shiftKey)return;event.preventDefault();items=[ref];setPrimary(ref);paint(stage);controls();edit(ref,box)};
   box.onpointerdown=event=>{
    if(event.button!==0||event.target.closest('[contenteditable="plaintext-only"]'))return;
    event.preventDefault();const additive=event.ctrlKey||event.metaKey||event.shiftKey;pick(ref,additive);paint(stage);controls();box.focus({preventScroll:true});
    if(additive){focus(stage);return;}
    if(window.SGL592?.locked622())return;const prior=before(),f=data(ref);
    if(items.length===1&&ref.kind==='image'&&event.target.closest('[data-transform]'))return MusicTransforms619.start(event,box,stage,f,paper,()=>history('',prior),()=>{redraw();controls()});
    const rect=stage.getBoundingClientRect(),sx=event.clientX,sy=event.clientY,original=positions(),limit=bounds(stage);let changed=false;box.setPointerCapture(event.pointerId);
    const move=ev=>{if(!changed&&Math.hypot(ev.clientX-sx,ev.clientY-sy)<3)return;if(!changed){history('',prior);changed=true}translate(stage,original,limit,(ev.clientX-sx)/rect.width*100,(ev.clientY-sy)/rect.height*100)};
    const end=()=>{box.removeEventListener('pointermove',move);box.removeEventListener('pointerup',end);box.removeEventListener('pointercancel',end);if(changed){redraw();controls();focus(stage)}};
    box.addEventListener('pointermove',move);box.addEventListener('pointerup',end);box.addEventListener('pointercancel',end);
   };
   box.onkeydown=event=>{
    if(event.target.closest('[contenteditable="plaintext-only"]'))return;
    if(event.key==='Enter'&&ref.kind==='field'){event.preventDefault();items=[ref];setPrimary(ref);paint(stage);controls();edit(ref,box);return}
    if(!['ArrowLeft','ArrowRight','ArrowUp','ArrowDown'].includes(event.key))return;if(window.SGL592?.locked622())return;
    event.preventDefault();if(!has(ref.kind,ref.index)){pick(ref);controls()}const d=event.shiftKey?1:.2;
    history('move:'+items.map(key).join(','));translate(stage,positions(),bounds(stage),event.key==='ArrowRight'?d:event.key==='ArrowLeft'?-d:0,event.key==='ArrowDown'?d:event.key==='ArrowUp'?-d:0);redraw();controls();focus(stage);
   };
  }
 }
 function remove(){if(window.SGL592?.locked622())return true;
  sync();const chosen=items.filter(ordinary);if(!chosen.length)return false;
  history();for(const kind of ['field','image'])for(const r of chosen.filter(r=>r.kind===kind).sort((a,b)=>b.index-a.index))(kind==='field'?model().fields:model().images610).splice(r.index,1);
  setPrimary({kind:'field',index:0});reset();redraw();controls();return true;
 }
 function copy(){sync();const chosen=items.filter(ordinary);if(!chosen.length)return false;clipboard=chosen.map(r=>({kind:r.kind,value:structuredClone(data(r))}));toast(chosen.length===1?'Elemento copiado.':'Elementos copiados.');return true}
 function paste(){if(window.SGL592?.locked622())return true;
  if(!clipboard.length)return false;const t=model();
  if(t.fields.length+clipboard.filter(r=>r.kind==='field').length>60||t.images610.length+clipboard.filter(r=>r.kind==='image').length>10){toast('La copia supera el límite de 60 textos o 10 imágenes.');return true}
  history();const list=clipboard.map(r=>({kind:r.kind,value:structuredClone(r.value)}));
  // A common offset preserves the relative arrangement, including rotated logos.
  let dx=2,dy=2;for(const r of list){const f=r.value;let right=f.x+f.width,bottom=f.y+(r.kind==='image'?f.height:0);if(r.kind==='image'){const a=(f.rotation619||0)*Math.PI/180,bw=Math.abs(f.width*Math.cos(a))+Math.abs(f.height*window.SGL592.paper().h/window.SGL592.paper().w*Math.sin(a)),bh=Math.abs(f.height*Math.cos(a))+Math.abs(f.width*window.SGL592.paper().w/window.SGL592.paper().h*Math.sin(a));right=f.x+f.width/2+bw/2;bottom=f.y+f.height/2+bh/2}dx=Math.max(0,Math.min(dx,100-right));dy=Math.max(0,Math.min(dy,95-bottom))}
  items=list.map(({kind,value})=>{value.x+=dx;value.y+=dy;const target=kind==='field'?t.fields:t.images610;target.push(value);return {kind,index:target.length-1}});setPrimary(items.at(-1));redraw();controls();focus(document.getElementById('music592_stage'));return true;
 }
 function duplicate(){sync();const chosen=items.filter(ordinary);if(!chosen.length)return false;const previous=clipboard;clipboard=chosen.map(r=>({kind:r.kind,value:structuredClone(data(r))}));const result=paste();clipboard=previous;return result}
 return {reset,sync,snapshot,restore,has,pick,paint,bind,remove,copy,paste,duplicate,count:()=>{sync();return items.length}};
};
