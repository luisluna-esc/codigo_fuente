/* Proportional corner resizing and clockwise rotation, shared with PDF layout. */
(()=>{
'use strict';
const rad=a=>a*Math.PI/180,deg=a=>a*180/Math.PI;
const rotate=(x,y,a)=>({x:x*Math.cos(a)-y*Math.sin(a),y:x*Math.sin(a)+y*Math.cos(a)});
const normal=a=>((a+180)%360+360)%360-180;
function fit(image,paper){
 const f={...image},a=rad(f.rotation619||0),w=f.width*paper.w/100,h=f.height*paper.h/100;
 const bw=Math.abs(w*Math.cos(a))+Math.abs(h*Math.sin(a)),bh=Math.abs(w*Math.sin(a))+Math.abs(h*Math.cos(a));
 const scale=Math.min(1,paper.w/bw,paper.h/bh);
 f.width*=scale;f.height*=scale;
 const halfW=bw*scale/2,halfH=bh*scale/2;
 const cx=Math.max(halfW,Math.min(paper.w-halfW,(image.x+image.width/2)*paper.w/100)),cy=Math.max(halfH,Math.min(paper.h-halfH,(image.y+image.height/2)*paper.h/100));
 f.x=cx/paper.w*100-f.width/2;f.y=cy/paper.h*100-f.height/2;
 return f;
}
function handles(){return ['nw','ne','sw','se','rotate'].map(h=>`<button type="button" data-transform="${h}" class="music619-handle music619-${h}" aria-label="${h==='rotate'?'Girar imagen':'Cambiar tamaño desde la esquina '+({nw:'superior izquierda',ne:'superior derecha',sw:'inferior izquierda',se:'inferior derecha'}[h])}" title="${h==='rotate'?'Girar imagen':'Cambiar tamaño conservando proporción'}">${h==='rotate'?'↻':''}</button>`).join('')}
function start(event,box,stage,image,paper,remember,finish){
 event.preventDefault();event.stopPropagation();const handle=event.target.closest('[data-transform]')?.dataset.transform;if(!handle)return;
 const original={...image},rect=stage.getBoundingClientRect(),a=rad(original.rotation619||0),w=original.width*paper.w/100,h=original.height*paper.h/100;
 const center={x:(original.x+original.width/2)*paper.w/100,y:(original.y+original.height/2)*paper.h/100};
 const point=e=>({x:(e.clientX-rect.left)/rect.width*paper.w,y:(e.clientY-rect.top)/rect.height*paper.h});
 const initial=point(event),angle=Math.atan2(initial.y-center.y,initial.x-center.x);
 const sx=handle.includes('e')?1:-1,sy=handle.includes('s')?1:-1,offset=rotate(-sx*w/2,-sy*h/2,a),anchor={x:center.x+offset.x,y:center.y+offset.y};
 let changed=false;box.setPointerCapture(event.pointerId);
 const move=e=>{
  const p=point(e);let next;
  if(handle==='rotate')next=fit({...original,rotation619:normal((original.rotation619||0)+deg(Math.atan2(p.y-center.y,p.x-center.x)-angle))},paper);
  else{
   const local=rotate(p.x-anchor.x,p.y-anchor.y,-a),minScale=Math.max(1/original.width,1/original.height);
   const scale=Math.max(minScale,(sx*local.x*w+sy*local.y*h)/(w*w+h*h)),nw=w*scale,nh=h*scale,c=rotate(sx*nw/2,sy*nh/2,a);
   next=fit({...original,width:original.width*scale,height:original.height*scale,x:(anchor.x+c.x-nw/2)/paper.w*100,y:(anchor.y+c.y-nh/2)/paper.h*100},paper);
  }
  if(!changed){remember();changed=true}Object.assign(image,next);
  Object.assign(box.style,{left:image.x+'%',top:image.y+'%',width:image.width+'%',height:image.height+'%',transform:`rotate(${image.rotation619||0}deg)`});
 };
 const end=()=>{box.removeEventListener('pointermove',move);box.removeEventListener('pointerup',end);box.removeEventListener('pointercancel',end);finish()};
 box.addEventListener('pointermove',move);box.addEventListener('pointerup',end);box.addEventListener('pointercancel',end);
}
window.MusicTransforms619={fit,handles,start,normal};
})();
