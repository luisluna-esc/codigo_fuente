'use strict';
(()=>{
 const modes={inline:'En línea',square:'Cuadrado',tight:'Estrecho',through:'Transparente',topbottom:'Arriba y abajo',behind:'Detrás del texto',front:'Encima del texto'};
 const editor=()=>document.getElementById('req_text');
 const submission=()=>v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);
 const editable=()=>!!editor()&&typeof v3CanEditCurrentRequirementDraft==='function'&&v3CanEditCurrentRequirementDraft();
 let selected=null,watermarkKey='',watermarkURL='';const anchors=new WeakMap();
 function setStyle(el,key,value){if(el.style.getPropertyValue(key)!==value||el.style.getPropertyPriority(key)!=='important')el.style.setProperty(key,value,'important');}
 // Public options are deliberately limited to the two requested modes. Legacy
 // layouts remain readable so opening an old report does not change its photos.
 const choices={through:'Transparente',front:'Encima del texto'};
 const contours=new WeakMap();
 function alphaBands(img){
  if(!img?.complete||!img.naturalWidth)return null;
  const key=img.currentSrc||img.src,cached=contours.get(img);if(cached?.key===key)return cached.bands;
  try{const c=document.createElement('canvas');c.width=64;c.height=64;const ctx=c.getContext('2d',{willReadFrequently:true});ctx.drawImage(img,0,0,64,64);const pixels=ctx.getImageData(0,0,64,64).data,bands=[];
   for(let y=0;y<64;y++){let first=64,last=-1;for(let x=0;x<64;x++)if(pixels[(y*64+x)*4+3]>12){first=Math.min(first,x);last=x;}
    if(last>=0)bands.push({left:first/64,right:(last+1)/64,top:y/64,bottom:(y+1)/64});}
   contours.set(img,{key,bands});return bands;
  }catch{return null;}
 }
 function wrapPolygon(rectangles,width){
  const rects=rectangles.filter(r=>r.right>r.left&&r.bottom>r.top),height=Math.max(1,...rects.map(r=>r.bottom));
  const area=rects.reduce((a,r)=>a+(r.right-r.left)*(r.bottom-r.top),0);
  const center=area?rects.reduce((a,r)=>a+(r.left+r.right)/2*(r.right-r.left)*(r.bottom-r.top),0)/area:width/2;
  const side=center>=width/2?'right':'left',outer=side==='right'?width:0;
  const cuts=[...new Set([0,height,...rects.flatMap(r=>[Math.max(0,r.top),r.bottom])])].sort((a,b)=>a-b),points=[[outer,0]];
  for(let i=0;i<cuts.length-1;i++){const a=cuts[i],b=cuts[i+1],active=rects.filter(r=>r.top<b&&r.bottom>a);
   const edge=active.length?(side==='right'?Math.min(...active.map(r=>r.left)):Math.max(...active.map(r=>r.right))):outer;
   points.push([Math.max(0,Math.min(width,edge)),a],[Math.max(0,Math.min(width,edge)),b]);}
  points.push([outer,height]);return {side,height,shape:'polygon('+points.map(([x,y])=>(x/width*100).toFixed(3)+'% '+(y/height*100).toFixed(3)+'%').join(',')+')'};
 }
 window.sgl548WrapPolygon=wrapPolygon;
 function refreshWrap(body,preview=false){
  if(!body)return;
  const figures=[...body.children].filter(n=>n.tagName==='FIGURE'&&n.dataset.sglLayout==='through');
  let mask=[...body.children].find(n=>n.classList.contains('sgl548-wrap-exclusion'));
  if(!figures.length){mask?.remove();return;}
  const width=Math.max(1,body.clientWidth||Number(figures[0].dataset.sglWrapHostWidth)||600),rectangles=[];
  for(const f of figures){const scale=Number(f.dataset.sglScale)||70,img=f.querySelector('img'),w=width*scale/100,aspect=Number(f.dataset.sglAspect)||((img?.naturalWidth||1)/(img?.naturalHeight||1)),h=w/Math.max(.05,aspect),left=Math.max(0,Number(f.dataset.sglLeft)||0),top=Math.max(0,Number(f.dataset.sglTop)||0),angle=(Number(f.dataset.sglAngle)||0)*Math.PI/180;
   const bands=Math.abs(angle)<.0001?alphaBands(img):null;
   if(bands){for(const b of bands)rectangles.push({left:Math.max(0,left+b.left*w-6),right:Math.min(width,left+b.right*w+6),top:Math.max(0,top+b.top*h-2),bottom:top+b.bottom*h+2});}
   else{const rw=Math.abs(w*Math.cos(angle))+Math.abs(h*Math.sin(angle)),rh=Math.abs(h*Math.cos(angle))+Math.abs(w*Math.sin(angle));rectangles.push({left:Math.max(0,left+(w-rw)/2-6),right:Math.min(width,left+(w+rw)/2+6),top:Math.max(0,top+(h-rh)/2-6),bottom:top+(h+rh)/2+6});}
  }
  if(!rectangles.length){mask?.remove();return;}
  const pageHeight=body.clientHeight||Infinity;for(const r of rectangles)r.bottom=Math.min(pageHeight,r.bottom);
  const geometry=wrapPolygon(rectangles,width);
  if(!mask){mask=document.createElement('span');mask.className='sgl548-wrap-exclusion';if(!preview)mask.dataset.sglEditorUi='1';mask.contentEditable='false';mask.setAttribute('aria-hidden','true');body.prepend(mask);}
  for(const [key,value] of Object.entries({display:'block',float:geometry.side,clear:'none',width:'100%',height:geometry.height+'px',margin:'0',padding:'0',border:'0','pointer-events':'none','shape-outside':geometry.shape,'shape-margin':'0'}))setStyle(mask,key,value);
 }
 window.sgl548RefreshWrap=refreshWrap;
 function throughLayout(figure,preview,geometry){
  const body=figure.closest('.sgl5039-page-body,.page-body')||figure.parentElement;if(!body)return false;
  const image=figure.querySelector('img'),width=Math.max(1,body.clientWidth||Number(geometry.hostWidth)||Number(figure.dataset.sglWrapHostWidth)||600);
  const aspect=Number(geometry.aspect)||Number(figure.dataset.sglAspect)||((image?.naturalWidth||1)/(image?.naturalHeight||1));
  const scale=Math.max(1,Math.min(100,Number(figure.dataset.sglScale)||70)),w=width*scale/100,h=w/Math.max(.05,aspect),bodyHeight=body.clientHeight||Number(geometry.hostHeight)||10000;
  const left=Math.max(0,Math.min(width-w,Number(figure.dataset.sglLeft)||0)),top=Math.max(0,Math.min(Math.max(0,bodyHeight-h),Number(figure.dataset.sglTop)||0));
  figure.dataset.sglFree='1';figure.dataset.sglScale=String(scale);figure.dataset.sglAspect=String(aspect);figure.dataset.sglLeft=String(left);figure.dataset.sglTop=String(top);figure.dataset.sglWrapHostWidth=String(width);
  const page=body.closest('.sgl5039-word-page');if(page)figure.dataset.sglPage=page.dataset.sgl5039Page;
  for(const [key,value] of Object.entries({position:'absolute',display:'block',float:'none',clear:'none',width:preview?scale+'%':w+'px','max-width':'100%',height:'auto',margin:'0',left:preview?left/width*100+'%':left+'px',top:top+'px','z-index':String(Math.max(1,Number(figure.dataset.sglZ)||1)),'shape-outside':'none','box-sizing':'border-box','aspect-ratio':aspect+' / 1'}))setStyle(figure,key,value);
  if(image){setStyle(image,'width','100%');setStyle(image,'height','auto');setStyle(image,'object-fit','contain');setStyle(image,'max-height','none');}
  delete figure.dataset.sglHeight;
  refreshWrap(body,preview);return true;
 }

 function layout(figure,preview=false,geometry={}){
  for(const [key,value] of Object.entries(geometry))if(['left','top','scale'].includes(key)&&value!==null&&value!==undefined)figure.dataset[{left:"sglLeft",top:"sglTop",scale:"sglScale"}[key]]=String(value);
  const mode=figure.dataset.sglLayout;if(!modes[mode])return false;
  const image=figure.querySelector('img');if(image)setStyle(image,'transform','rotate('+(Number(figure.dataset.sglAngle)||0)+'deg)');
  if(!preview&&figure.dataset.sglAnchor==='text'&&['front','behind','through'].includes(mode)){
   const body=figure.closest('.sgl5039-page-body'),blocks=[...body?.children||[]].filter(n=>n.tagName!=='FIGURE'&&!n.hasAttribute('data-sgl-editor-ui'));
   let anchor=anchors.get(figure);if(!anchor?.isConnected){anchor=blocks[Number(figure.dataset.sglAnchorIndex)||0];if(anchor)anchors.set(figure,anchor);}
   if(anchor){if(geometry.top!==null&&geometry.top!==undefined)figure.dataset.sglAnchorOffset=String(Number(geometry.top)-anchor.offsetTop);const host=anchor.closest('.sgl5039-page-body');if(host&&host!==body){host.append(figure);figure.dataset.sglPage=host.closest('.sgl5039-word-page').dataset.sgl5039Page;}
    figure.dataset.sglTop=String(Math.max(0,anchor.offsetTop+(Number(figure.dataset.sglAnchorOffset)||0)));figure.dataset.sglAnchorIndex=String([...host.children].filter(n=>n.tagName!=='FIGURE'&&!n.hasAttribute('data-sgl-editor-ui')).indexOf(anchor));}
  }
  if(mode==='through')return throughLayout(figure,preview,geometry);
  refreshWrap(figure.closest('.sgl5039-page-body,.page-body')||figure.parentElement,preview);
  if(mode==='front'){setStyle(figure,'shape-outside','none');setStyle(figure,'display','block');return false;}
  const free=mode==='behind';figure.dataset.sglFree=free?'1':'0';
  setStyle(figure,'position',free?'absolute':'relative');setStyle(figure,'display',mode==='inline'?'inline-block':'block');
  setStyle(figure,'float',['square','tight','through'].includes(mode)?'left':'none');setStyle(figure,'clear',mode==='topbottom'?'both':'none');
  setStyle(figure,'width',Math.max(1,Math.min(100,Number(figure.dataset.sglScale)||70))+'%');setStyle(figure,'max-width','100%');
  setStyle(figure,'z-index',free?'-1':'auto');setStyle(figure,'margin',free?'0':'6px 12px 6px 0');
  setStyle(figure,'left',free?(Number(figure.dataset.sglLeft)||0)+'px':'auto');setStyle(figure,'top',free?(Number(figure.dataset.sglTop)||0)+'px':'auto');
  const source=figure.querySelector('img')?.src||'';
  const contour=['tight','through'].includes(mode)&&/^data:image\/(png|webp|jpeg);base64,[A-Za-z0-9+/=]+$/.test(source);
  setStyle(figure,'shape-outside',contour?'url("'+source+'")':'none');setStyle(figure,'shape-margin',mode==='tight'?'8px':'0px');setStyle(figure,'shape-image-threshold',mode==='through'?'.05':'.5');
  if(preview)figure.parentElement?.style.setProperty('isolation','isolate');
  return true;
 }
 window.sgl548ApplyPhotoLayout=layout;
 function dirty(){if(!editable())return;submission().updatedAt=isoNow();window.sgl5047MarkReportDirty?.();editor().dispatchEvent(new Event('input',{bubbles:true}));}
 function choose(mode){
  if(!editable()||!selected?.isConnected||!choices[mode])return;
  selected.dataset.sglLayout=mode;selected.dataset.sglFree=['front','behind','through'].includes(mode)?'1':'0';
  if(mode==='front')window.sgl5045ApplyStablePhoto?.(selected);else layout(selected);
  const id=selected.dataset.sglPhotoId;submission().reportPhotoBoxes||={};submission().reportPhotoBoxes[id]={...(submission().reportPhotoBoxes[id]||{}),layout:mode,free:['front','behind','through'].includes(mode)};
  dirty();document.querySelector('.sgl548-layout-menu')?.remove();
 }
 function modeIcon(key){const lines=key==='through'?'<path d="M3 5h26M3 11h5m18 0h3M3 17h5m18 0h3M3 23h5m18 0h3M3 29h26"/>':'<path d="M3 5h26M3 11h26M3 17h26M3 23h26M3 29h26"/>';return '<svg viewBox="0 0 32 32" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.5">'+lines+'<rect x="11" y="10" width="12" height="15" rx="1" fill="var(--surface,#fff)"/></svg>';}

 function anchorMode(mode){if(!selected||!editable())return;selected.dataset.sglAnchor=mode;
  const body=selected.closest('.sgl5039-page-body'),blocks=[...body.children].filter(n=>n.tagName!=='FIGURE'&&!n.hasAttribute('data-sgl-editor-ui'));const anchor=blocks.filter(n=>n.offsetTop<=Number(selected.dataset.sglTop||0)).at(-1)||blocks[0];
  if(anchor){anchors.set(selected,anchor);selected.dataset.sglAnchorIndex=String(blocks.indexOf(anchor));selected.dataset.sglAnchorOffset=String(Number(selected.dataset.sglTop||0)-anchor.offsetTop);}dirty();}
 function openLayout(button){
  if(!selected?.isConnected)return toast('Selecciona una imagen del informe.');
  document.querySelector('.sgl548-layout-menu')?.remove();const panel=document.createElement('div');panel.className='sgl548-layout-menu';panel.setAttribute('role','dialog');panel.setAttribute('aria-label','Opciones de imagen');
  panel.innerHTML='<h4>Opciones de imagen</h4><div class="sgl548-layout-options">'+Object.entries(choices).map(([key,label])=>`<button type="button" data-mode="${key}" aria-pressed="${(selected.dataset.sglLayout||'front')===key}">${modeIcon(key)}<span>${label}</span></button>`).join('')+'</div><p class="muted" style="font-size:11px">Transparente: el texto rodea la imagen. Encima del texto: la imagen se superpone. Arrastra para moverla.</p><button type="button" class="btn ghost small" data-close>Cerrar</button>';
  const rotations=document.createElement('div');rotations.className='actions';rotations.innerHTML='<button type="button" class="btn ghost" data-angle="-90">↶ Girar 90°</button><button type="button" class="btn ghost" data-angle="90">↻ Girar 90°</button>';panel.querySelector('[data-close]').before(rotations);rotations.querySelectorAll('[data-angle]').forEach(b=>b.onclick=()=>{if(!editable()||!selected?.isConnected)return;selected.dataset.sglAngle=String(((Number(selected.dataset.sglAngle)||0)+Number(b.dataset.angle)+360)%360);layout(selected);dirty();});
  panel.querySelectorAll('[data-mode]').forEach(b=>b.onclick=()=>choose(b.dataset.mode));panel.querySelector('[data-close]').onclick=()=>panel.remove();
  const rect=button.getBoundingClientRect();panel.style.left=Math.max(8,Math.min(innerWidth-276,rect.left))+'px';panel.style.top=Math.max(8,Math.min(innerHeight-230,rect.top-225))+'px';document.body.appendChild(panel);const bounds=panel.getBoundingClientRect();panel.style.top=Math.max(8,Math.min(innerHeight-bounds.height-8,rect.top-bounds.height-8))+'px';
 }
 async function paintWatermark(){
  const root=editor();if(!root)return;const mark=window.sgl549WatermarkFor?.(modalOwnerId);
  if(!mark){root.querySelectorAll('.sgl548-watermark').forEach(n=>n.remove());return;}
  watermarkURL=mark.dataUrl||'';
  if(!/^data:image\/png;base64,[A-Za-z0-9+/=]+$/.test(watermarkURL))return;
  root.querySelectorAll(':scope > .sgl5039-word-page').forEach((page,index)=>{let layer=page.querySelector(':scope > .sgl548-watermark');if(!mark.allPages&&index+1!==mark.page){layer?.remove();return;}if(!layer){layer=document.createElement('div');layer.className='sgl548-watermark';layer.dataset.sglEditorUi='1';layer.contentEditable='false';layer.setAttribute('aria-hidden','true');layer.append(document.createElement('img'));page.prepend(layer);}const img=layer.firstElementChild;if(img.src!==watermarkURL)img.src=watermarkURL;img.style.width=mark.size+'%';img.style.opacity=String(mark.opacity);});
 }
 function decorateRotation(figure){
  let handle=figure.querySelector(':scope > .sgl548-rotation');if(!editable()){handle?.remove();return;}const fit=()=>{const mobile=matchMedia('(max-width:820px)').matches,zoom=Math.max(.2,figure.getBoundingClientRect().width/Math.max(1,figure.offsetWidth));for(const [key,value] of Object.entries({width:(mobile?36/zoom:30)+'px',height:(mobile?36/zoom:30)+'px',top:(mobile?-44/zoom:-34)+'px','font-size':(mobile?18/zoom:16)+'px'}))setStyle(handle,key,value);};if(handle){fit();return;}
  handle=document.createElement('button');handle.type='button';handle.className='sgl548-rotation';handle.dataset.sglEditorUi='1';handle.contentEditable='false';handle.title='Girar imagen';handle.setAttribute('aria-label','Girar imagen');handle.textContent='↻';figure.append(handle);fit();
  handle.addEventListener('pointerdown',event=>{event.preventDefault();event.stopImmediatePropagation();selected=figure;handle.setPointerCapture(event.pointerId);const rect=figure.getBoundingClientRect(),cx=rect.left+rect.width/2,cy=rect.top+rect.height/2,start=Math.atan2(event.clientY-cy,event.clientX-cx),angle=Number(figure.dataset.sglAngle)||0;
   const move=e=>{figure.dataset.sglAngle=String(Math.round((angle+(Math.atan2(e.clientY-cy,e.clientX-cx)-start)*180/Math.PI+360)%360));layout(figure);};
   const end=()=>{handle.removeEventListener('pointermove',move);handle.removeEventListener('pointerup',end);handle.removeEventListener('pointercancel',end);dirty();};handle.addEventListener('pointermove',move);handle.addEventListener('pointerup',end);handle.addEventListener('pointercancel',end);
  });
  handle.addEventListener('keydown',e=>{if(!['ArrowLeft','ArrowRight','Home'].includes(e.key))return;e.preventDefault();figure.dataset.sglAngle=String(e.key==='Home'?0:((Number(figure.dataset.sglAngle)||0)+(e.key==='ArrowLeft'?-5:5)+360)%360);layout(figure);dirty();});
 }
 document.addEventListener('load',e=>{if(e.target.tagName==='IMG'&&e.target.closest?.('#req_text figure[data-sgl-photo-id]'))requestAnimationFrame(enhance);},true);
 document.addEventListener('input',e=>{if(e.target.closest?.('#req_text'))requestAnimationFrame(enhance);});
 let scheduled=false;
 function decorateOptions(figure){
  let b=figure.querySelector(':scope > .sgl548-photo-options');if(!editable()){b?.remove();return;}
  if(!b){b=document.createElement('button');b.type='button';b.className='sgl548-photo-options';b.dataset.sglEditorUi='1';b.contentEditable='false';b.title='Opciones de imagen';b.setAttribute('aria-label','Opciones de imagen');b.textContent='▧';b.addEventListener('pointerdown',e=>{e.preventDefault();e.stopPropagation();selected=figure;openLayout(b);});b.onclick=e=>{e.preventDefault();e.stopPropagation();if(!document.querySelector('.sgl548-layout-menu')){selected=figure;openLayout(b);}};figure.append(b);}
  const zoom=Math.max(.1,figure.getBoundingClientRect().width/Math.max(1,figure.offsetWidth));for(const [k,v] of Object.entries({width:32/zoom+'px',height:32/zoom+'px',top:-40/zoom+'px','font-size':18/zoom+'px'}))setStyle(b,k,v);
 }
 function enhance(){
  const root=editor();if(!root)return;
  const modal=root.closest('.modal'),bar=modal.querySelector('.sgl537-page-actions'),controls=modal.querySelector('.sgl532-controls'),desktop=matchMedia('(min-width:821px)').matches;
  if(bar&&controls&&desktop&&bar.parentElement!==controls)controls.prepend(bar);
  else if(bar&&!desktop&&controls?.contains(bar))modal.querySelector('.sgl5037-paper-toolbar')?.after(bar);
  for(const b of bar?.querySelectorAll('button')||[]){if(/sgl537CreatePage/.test(b.getAttribute('onclick')||'')||b.getAttribute('aria-label')==='Crear nueva página'){b.title='Crear página';b.setAttribute('aria-label','Crear página');if(b.textContent!=='▱+')b.textContent='▱+';}}
  modal.querySelectorAll('[data-watermark548]').forEach(b=>b.remove());
  controls?.querySelectorAll('[data-layout548]').forEach(b=>b.remove());
  for(const f of root.querySelectorAll('figure[data-sgl-photo-id]')){if(!f.dataset.sglLayout){const saved=submission().reportPhotoBoxes?.[f.dataset.sglPhotoId];f.dataset.sglLayout=saved?.layout||(saved?(saved.free?'front':'topbottom'):'front');if(!saved){f.dataset.sglFree='1';window.sgl5045ApplyStablePhoto?.(f);}}const saved=submission().reportPhotoBoxes?.[f.dataset.sglPhotoId];if(!f.dataset.sglAngle)f.dataset.sglAngle=String(saved?.angle||0);if(!f.dataset.sglAnchor){f.dataset.sglAnchor=saved?.anchor||'page';f.dataset.sglAnchorOffset=String(saved?.anchorOffset||0);f.dataset.sglAnchorIndex=String(saved?.anchorIndex||0);}if(f.dataset.sglAnchorOffset===undefined)f.dataset.sglAnchorOffset=String(saved?.anchorOffset||0);if(f.dataset.sglAnchorIndex===undefined)f.dataset.sglAnchorIndex=String(saved?.anchorIndex||0);layout(f);if(f.dataset.sglLayout==='front')window.sgl5045ApplyStablePhoto?.(f);decorateRotation(f);decorateOptions(f);}
  root.querySelectorAll('.sgl5039-page-body').forEach(body=>refreshWrap(body));
  paintWatermark().catch(()=>{});
 }
 window.addEventListener('pointerdown',e=>{const figure=e.target.closest?.('#req_text figure[data-sgl-photo-id]');if(figure){selected=figure;editor()?.querySelectorAll('.sgl548-photo-selected').forEach(f=>f.classList.remove('sgl548-photo-selected'));figure.classList.add('sgl548-photo-selected');}else if(!e.target.closest?.('.sgl548-layout-menu')){editor()?.querySelectorAll('.sgl548-photo-selected').forEach(f=>f.classList.remove('sgl548-photo-selected'));document.querySelector('.sgl548-layout-menu[aria-label="Opciones de imagen"]')?.remove();}},true);
 new MutationObserver(()=>{if(scheduled)return;scheduled=true;requestAnimationFrame(()=>{scheduled=false;enhance();});}).observe(document.body,{childList:true,subtree:true});
 window.sgl548EnhanceEditor=enhance;
 window.addEventListener('resize',enhance);
})();
