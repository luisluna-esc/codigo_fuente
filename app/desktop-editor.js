/* SGL 5.47: desktop report controls. Native transport and mobile editor stay independent. */
(function(){
  'use strict';
  const css=document.createElement('link');css.rel='stylesheet';css.href='desktop-editor.css?v=5.48-r1';document.head.appendChild(css);
  const states=new WeakMap(),desktop=()=>window.matchMedia('(min-width:821px)').matches;
  const root=()=>document.getElementById('req_text');
  const context=()=>[currentUser()?.id,modalOwnerId,v3ModalProgramId,modalRequirementId].join('|');
  function state(){const editor=root();if(!editor)return null;let s=states.get(editor);if(!s){s={editor,key:context(),bookmark:null,photo:'',busy:false,history:[],redo:[],lastHtml:html(),applyingHistory:false};states.set(editor,s);}return s;}
  function canEdit(s){return !!(s?.editor.isConnected&&s.key===context()&&v3CanEditCurrentRequirementDraft?.()&&s.editor.querySelector('[contenteditable="true"]'));}
  function bodies(s){return [...s.editor.querySelectorAll('.sgl5039-page-body')];}
  function textNodes(s){const out=[],walk=document.createTreeWalker(s.editor,NodeFilter.SHOW_TEXT);let n;while(n=walk.nextNode())if(n.parentElement?.closest('.sgl5039-page-body')&&!n.parentElement.closest('figure,[data-sgl-editor-ui],[data-sgl5039-static]'))out.push(n);return out;}
  function offset(s,node,at){let total=0;const before=document.createRange();before.selectNodeContents(s.editor);try{before.setEnd(node,at);}catch{return null;}
    for(const n of textNodes(s)){if(n===node)return total+at;if(before.comparePoint(n,0)>0)break;if(before.comparePoint(n,n.length)<=0)total+=n.length;}
    return total;
  }
  function capture(force=false){const s=state(),sel=getSelection();if(!s||!desktop()||!sel?.rangeCount||(s.toolbarActive&&force!==true))return;const r=sel.getRangeAt(0),inBody=n=>(n.nodeType===1?n:n.parentElement)?.closest('.sgl5039-page-body');
    if(!s.editor.contains(r.startContainer)||!s.editor.contains(r.endContainer)||!inBody(r.startContainer)||!inBody(r.endContainer))return;
    s.bookmark={start:offset(s,r.startContainer,r.startOffset),end:offset(s,r.endContainer,r.endOffset),live:r.cloneRange()};
    const first=!r.collapsed&&selectedText(s,r)[0]?.n,el=first?.parentElement||(r.startContainer.nodeType===1?r.startContainer:r.startContainer.parentElement),style=getComputedStyle(el),bar=s.editor.closest('.modal')?.querySelector('.sgl532-controls');
    if(bar&&document.activeElement?.closest('.sgl532-controls')===null){
      const size=Math.round(parseFloat(style.fontSize)*.75),font=/times/i.test(style.fontFamily)?'Times New Roman':'Arial';
      if([10,11,12,13,14].includes(size))bar.querySelector('[data-format="size"]').value=String(size);
      bar.querySelector('[data-format="font"]').value=font;
      for(const [cmd,on] of [['bold',Number(style.fontWeight)>=600],['italic',style.fontStyle==='italic'],['underline',style.textDecorationLine.includes('underline')]])bar.querySelector(`[data-command="${cmd}"]`)?.setAttribute('aria-pressed',String(on));
    }
  }
  function locate(s,value,start=false){let left=Math.max(0,value||0),last=null;for(const n of textNodes(s)){last=n;if(left<n.length||left===n.length&&!start)return [n,left];left-=n.length;}return last?[last,last.length]:[bodies(s)[0],0];}
  function restore(s,bookmark=s.bookmark){if(!canEdit(s))return null;const b=bookmark||{start:0,end:0},a=locate(s,b.start,true),z=locate(s,b.end);if(!a[0]||!z[0])return null;
    const body=(a[0].nodeType===1?a[0]:a[0].parentElement).closest('.sgl5039-page-body');if(b.start===b.end)body?.focus({preventScroll:true});else document.activeElement?.blur();window.__sgl5039ActiveBody=body;
    const live=bookmark?.live,inBody=n=>(n?.nodeType===1?n:n?.parentElement)?.closest('.sgl5039-page-body');
    const usable=live&&inBody(live.startContainer)&&inBody(live.endContainer)&&s.editor.contains(live.startContainer)&&s.editor.contains(live.endContainer)&&(!live.collapsed||b.start===b.end);
    const r=usable?live.cloneRange():document.createRange();if(!usable){r.setStart(...a);r.setEnd(...z);}const sel=getSelection();sel.removeAllRanges();sel.addRange(r);return r;
  }
  function html(){return readRichEditor().html;}
  function recordHistory(s){
    if(!canEdit(s)||s.applyingHistory)return;
    const after=html();if(after===s.lastHtml)return;
    if(s.busy){s.lastHtml=after;s.history=[];s.redo=[];return;}
    s.history.push({before:s.lastHtml,after,bookmark:{...s.bookmark}});if(s.history.length>60)s.history.shift();s.redo=[];s.lastHtml=after;
    for(const button of s.editor.closest('.modal').querySelectorAll('[data-history]'))button.disabled=!(button.dataset.history==='undo'?s.history:s.redo).length;
  }
  function changed(s,before){s.editor.closest('.modal').__sgl5047Dirty=true;recordHistory(s);s.editor.dispatchEvent(new Event('input',{bubbles:true}));}
  function historyAction(s,action){
    if(!canEdit(s))return;recordHistory(s);
    const undo=action==='undo',from=undo?s.history:s.redo,to=undo?s.redo:s.history,last=from.at(-1);if(!last)return;
    const value=undo?last.before:last.after;s.applyingHistory=true;
    bodies(s).forEach((body,i)=>body.innerHTML=i?'':value);s.bookmark=last.bookmark;from.pop();to.push(last);s.lastHtml=value;
    restore(s);changed(s);s.applyingHistory=false;window.sgl5039ScheduleReflow?.(0);window.sgl512HydrateInlinePhotos?.().catch?.(()=>{});enhance();
  }
  function selectedText(s,r){return textNodes(s).filter(n=>r.intersectsNode(n)).map(n=>({n,start:n===r.startContainer?r.startOffset:0,end:n===r.endContainer?r.endOffset:n.length})).filter(x=>x.end>x.start);}
  function applyStyle(property,value){const s=state();if(!canEdit(s))return;const r=restore(s);if(!r)return;const before=html(),saved={...s.bookmark};
    if(r.collapsed){
      // A styled insertion point also supports choosing the font before typing.
      const span=document.createElement('span');span.style.setProperty(property,value,property==='color'?'important':'');const n=document.createTextNode('\u200b');span.appendChild(n);r.insertNode(span);r.setStart(n,1);r.collapse(true);getSelection().removeAllRanges();getSelection().addRange(r);capture(true);
    }else{
      for(const {n,start,end} of selectedText(s,r).reverse()){
        if(end<n.length)n.splitText(end);const part=start?n.splitText(start):n,span=document.createElement('span');span.style.setProperty(property,value,property==='color'?'important':'');part.replaceWith(span);span.appendChild(part);
      }
      s.bookmark={...saved,live:null};restore(s);
    }
    changed(s,before);
  }
  // Typed/pasted text may be directly inside an editing host, without a paragraph.
  // Normalize only those inline runs before applying paragraph commands.
  const blockSelector='p,div,li,h1,h2,h3,h4,h5,h6,blockquote,pre';
  function selectedBlocks(s){
    let altered=false;
    for(const body of bodies(s)){
      let paragraph=null;
      for(const node of [...body.childNodes]){
        if(node.nodeType===Node.COMMENT_NODE)continue;
        if(node.nodeType===Node.ELEMENT_NODE&&node.matches('p,div,ul,ol,figure,table,h1,h2,h3,h4,h5,h6,blockquote,pre,[data-sgl-editor-ui]')){paragraph=null;continue;}
        if(!paragraph){paragraph=document.createElement('p');body.insertBefore(paragraph,node);altered=true;}
        paragraph.appendChild(node);if(node.nodeName==='BR')paragraph=null;
      }
      if(!body.childNodes.length){body.innerHTML='<p><br></p>';altered=true;}
    }
    if(altered&&s.bookmark)s.bookmark={...s.bookmark,live:null};
    const range=restore(s);if(!range)return [];
    return bodies(s).flatMap(body=>[...body.querySelectorAll(blockSelector)]).filter(n=>!n.closest('figure,[data-sgl-editor-ui]')&&!n.querySelector(blockSelector)&&range.intersectsNode(n));
  }
  function nativeCommand(command,value=null){const s=state();if(!canEdit(s))return;const r=restore(s);if(!r)return;const before=html(),saved={...s.bookmark};
    if(command.startsWith('justify')){const alignment={justifyLeft:'left',justifyCenter:'center',justifyRight:'right',justifyFull:'justify'}[command];
      const blocks=selectedBlocks(s);
      for(const block of blocks)block.style.textAlign=alignment;changed(s,before);return;
    }
    // Each page is its own editing host. Apply the command separately to each
    // selected host, then restore the logical selection across repagination.
    const hosts=bodies(s).filter(b=>r.intersectsNode(b));
    if(hosts.length<=1)document.execCommand(command,false,value);
    else for(const body of hosts){
      const part=document.createRange();part.selectNodeContents(body);
      if(body.contains(r.startContainer))part.setStart(r.startContainer,r.startOffset);
      if(body.contains(r.endContainer))part.setEnd(r.endContainer,r.endOffset);
      body.focus({preventScroll:true});getSelection().removeAllRanges();getSelection().addRange(part);document.execCommand(command,false,value);
    }
    s.bookmark={...saved,live:null};restore(s);changed(s,before);
  }
  function list(kind){
    const s=state();if(!canEdit(s))return;const range=restore(s);if(!range)return;const before=html(),saved={...s.bookmark},tag=kind==='decimal'?'OL':'UL';
    const blocks=selectedBlocks(s);
    const converted=new Set();let previous=null,group=null;
    for(const block of blocks){
      if(block.tagName==='LI'){
        const parent=block.parentElement;if(converted.has(parent))continue;converted.add(parent);
        if(parent.tagName===tag){if(tag==='UL')parent.style.listStyleType=kind;else{const paragraphs=[...parent.children].map(li=>{const p=document.createElement('p');p.innerHTML=li.innerHTML;p.style.cssText=li.style.cssText;return p;});parent.replaceWith(...paragraphs);}}
        else{const next=document.createElement(tag);next.style.listStyleType=kind;next.append(...parent.childNodes);parent.replaceWith(next);}
        continue;
      }
      const item=document.createElement('li');item.innerHTML=block.innerHTML;item.style.cssText=block.style.cssText;
      if(previous?.nextElementSibling===block&&group){group.appendChild(item);block.remove();}
      else{group=document.createElement(tag);group.style.listStyleType=kind;group.appendChild(item);block.replaceWith(group);}
      previous=group;
    }
    s.bookmark={...saved,live:null};restore(s);changed(s,before);
  }
  const icon=(paths)=>`<svg viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${paths}</svg>`;
  const alignIcon=mode=>icon([4,9,14,19].map((y,i)=>{const short=i%2,w=short?11:18,x=mode==='center'?(24-w)/2:mode==='right'?21-w:3;return `<path d="M${x} ${y}h${mode==='justify'?18:w}"/>`;}).join(''));
  function createBar(s,toolbar){const bar=document.createElement('div');bar.className='sgl532-controls';bar.setAttribute('role','group');bar.setAttribute('aria-label','Formato del informe');
    bar.innerHTML=`<div class="sgl532-group"><button type="button" data-history="undo" title="Deshacer (Ctrl+Z)" aria-label="Deshacer">↶</button><button type="button" data-history="redo" title="Rehacer (Ctrl+Y)" aria-label="Rehacer">↷</button></div><div class="sgl532-group"><button type="button" data-command="bold" title="Negrita (Ctrl+B)" aria-label="Negrita" aria-pressed="false"><b>N</b></button><button type="button" data-command="underline" title="Subrayado (Ctrl+U)" aria-label="Subrayado" aria-pressed="false"><u>S</u></button><button type="button" data-command="italic" title="Cursiva (Ctrl+I)" aria-label="Cursiva" aria-pressed="false"><i>C</i></button></div>
      <div class="sgl532-group">${[['left','Izquierda'],['center','Centrar'],['right','Derecha'],['justify','Justificar']].map(([m,label])=>`<button type="button" data-command="justify${m==='justify'?'Full':m[0].toUpperCase()+m.slice(1)}" aria-label="${label}" title="${label}">${alignIcon(m)}</button>`).join('')}</div>
      <div class="sgl532-group"><select data-format="font" aria-label="Tipo de fuente" title="Tipo de fuente"><option>Arial</option><option>Times New Roman</option></select><select data-format="size" aria-label="Tamaño de fuente" title="Tamaño de fuente">${[10,11,12,13,14].map(n=>`<option${n===12?' selected':''}>${n}</option>`).join('')}</select></div>
      <div class="sgl532-group"><label title="Color de fuente" class="sgl532-color">A<input type="color" value="#000000" data-format="color" aria-label="Color de fuente"></label><label title="Resaltado de texto" class="sgl532-color sgl532-highlight">▰<input type="color" value="#fff176" data-format="highlight" aria-label="Resaltado de texto"></label></div>
      <div class="sgl532-group"><select data-format="bullets" aria-label="Viñetas" title="Viñetas"><option value="">●</option><option value="disc">● Circular</option><option value="circle">○ Círculo vacío</option><option value="square">▪ Cuadrada</option></select><button type="button" data-list="decimal" title="Numeración" aria-label="Numeración">1.</button></div>
      <button type="button" class="sgl532-mic" data-mic title="Probar micrófono">${icon('<rect x="9" y="2" width="6" height="13" rx="3"/><path d="M5 11a7 7 0 0014 0M12 18v4M8 22h8"/>')}<span>Probar micrófono</span></button><button type="button" data-dictate class="sgl533-dictate" title="Dictar desde el cursor" aria-label="Dictar desde el cursor" aria-pressed="false">${icon('<rect x="9" y="2" width="6" height="13" rx="3"/><path d="M5 11a7 7 0 0014 0M12 18v4M8 22h8"/>')}</button><button type="button" data-delete-report class="sgl533-delete" title="Eliminar avance del informe" aria-label="Eliminar avance del informe" hidden>${icon('<path d="M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7M14 10v7"/>')}</button>`;
    bar.addEventListener('pointerdown',e=>{capture(true);s.toolbarActive=true;if(e.target.closest('button'))e.preventDefault();});
    bar.addEventListener('focusin',()=>{capture();s.toolbarActive=true;});
    bar.addEventListener('click',e=>{const b=e.target.closest('button');if(!b||b.disabled)return;if(b.dataset.history)return historyAction(s,b.dataset.history);if(b.dataset.turn)return rotate(Number(b.dataset.turn));if(b.hasAttribute('data-mic'))return window.sgl510TestMicrophone?.();if(b.hasAttribute('data-delete-report'))return s.editor.closest('.modal')?.querySelector('[data-sgl5031-delete-progress-head],.sgl5088-delete-progress-modal')?.click();if(b.hasAttribute('data-dictate')){restore(s);return s.editor.closest('.modal')?.querySelector('.sgl-leader-dictation-fab')?.click();}if(b.dataset.command)nativeCommand(b.dataset.command);if(b.dataset.list)list(b.dataset.list);});
    bar.addEventListener('change',e=>{const input=e.target,kind=input.dataset.format;if(!kind)return;
      if(kind==='size'&&['10','11','12','13','14'].includes(input.value))applyStyle('font-size',input.value+'pt');
      if(kind==='font'&&['Arial','Times New Roman'].includes(input.value))applyStyle('font-family',input.value);
      if(kind==='color')applyStyle('color',input.value);if(kind==='highlight')applyStyle('background-color',input.value);
      if(kind==='bullets'&&input.value){list(input.value);input.value='';}
    });
    toolbar.querySelector('.sgl5037-paper-zoom')?.before(bar);if(!bar.isConnected)toolbar.appendChild(bar);
    const actions=document.createElement('div');actions.className='sgl532-window-actions';actions.innerHTML=`<button type="button" class="sgl532-max" title="Maximizar informe" aria-label="Maximizar informe">${icon('<path d="M8 3H3v5M16 3h5v5M3 16v5h5M21 16v5h-5"/>')}</button><button type="button" class="sgl532-close" title="Cerrar informe" aria-label="Cerrar informe">${icon('<path d="M6 6l12 12M18 6L6 18"/>')}</button>`;
    actions.querySelector('.sgl532-max').onclick=e=>{window.sgl5034ToggleReportMaximize?.();const max=s.editor.closest('.modal').classList.contains('sgl5034-modal-maximized'),b=e.currentTarget;b.title=max?'Restaurar tamaño':'Maximizar informe';b.setAttribute('aria-label',b.title);b.setAttribute('aria-pressed',String(max));};
    actions.querySelector('.sgl532-close').onclick=()=>closeModal();
    for(const action of bar.querySelectorAll('[data-dictate],[data-delete-report]')){
      action.addEventListener('pointerdown',e=>{capture();e.preventDefault();});
      action.addEventListener('click',()=>{if(action.disabled)return;const modal=s.editor.closest('.modal');if(action.hasAttribute('data-dictate')){restore(s);modal.querySelector('.sgl-leader-dictation-fab')?.click();}else modal.querySelector('[data-sgl5031-delete-progress-head],.sgl5088-delete-progress-modal')?.click();});
      actions.insertBefore(action,actions.querySelector('.sgl532-max'));
    }
    toolbar.appendChild(actions);
  }
  function selectedPhoto(s){
    return [...s.editor.querySelectorAll('figure[data-sgl-photo-id]')].find(f=>f.dataset.sglPhotoId===s.photo)||s.editor.querySelector('.sgl5045-photo-selected')||null;
  }
  function rememberPhoto(figure){
    const s=state();if(!s)return;
    s.photo=figure&&s.editor.contains(figure)?String(figure.dataset.sglPhotoId||''):'';
    enhance();
  }
  window.sgl543RememberPhoto=rememberPhoto;
  function decoratePhotoTurns(s){s.editor.querySelectorAll('.sgl543-photo-turns').forEach(n=>n.remove());}

  async function rotate(direction,photoId=''){const s=state(),figure=s&&(photoId?[...s.editor.querySelectorAll('figure[data-sgl-photo-id]')].find(f=>f.dataset.sglPhotoId===photoId):selectedPhoto(s));if(!canEdit(s)||s.busy||!figure||![-1,1].includes(direction))return toast('Selecciona la foto que deseas girar.');
    s.busy=true;enhance();const oldId=figure.dataset.sglPhotoId;let created=null;
    try{
      const file=await fileGet(oldId);if(!file?.dataUrl||file.ownerId!==modalOwnerId)throw new Error('No se encontró la foto de este informe.');
      const img=new Image();await new Promise((resolve,reject)=>{img.onload=resolve;img.onerror=()=>reject(new Error('No se pudo leer la imagen.'));img.src=file.dataUrl;});
      const canvas=document.createElement('canvas');canvas.width=img.naturalHeight;canvas.height=img.naturalWidth;const ctx=canvas.getContext('2d');if(!ctx)throw new Error('El navegador no pudo preparar la foto.');
      ctx.translate(canvas.width/2,canvas.height/2);ctx.rotate(direction*Math.PI/2);ctx.drawImage(img,-img.naturalWidth/2,-img.naturalHeight/2);
      const dataUrl=canvas.toDataURL('image/png'),id=newId('file'),next={...file,id,dataUrl,type:'image/png',name:String(file.name||'Foto').replace(/\.[^.]+$/,'')+'.png',size:dataUrlBytes(dataUrl),width:canvas.width,height:canvas.height,reportAspect:canvas.width/canvas.height,createdAt:isoNow()};
      delete next.sha256;delete next.reportHeight;
      await filePut(next);created=id;if(!canEdit(s))throw new Error('La sesión o el informe cambió.');
      const live=[...s.editor.querySelectorAll('figure[data-sgl-photo-id]')].find(f=>f.dataset.sglPhotoId===oldId);if(!live)throw new Error('La foto ya no está en el informe.');
      const body=live.closest('.sgl5039-page-body'),bodyWidth=body?.clientWidth||627,maxHeight=340;
      next.reportScale=Math.min(Number(live.dataset.sglScale||file.reportScale||62),Math.max(1,Math.floor(maxHeight*next.reportAspect/bodyWidth*100)));
      await filePut(next);if(!canEdit(s)||!live.isConnected)throw new Error('El informe cambió antes de aplicar el giro.');
      const sub=v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);if(!sub.attachmentIds?.includes(oldId))throw new Error('La foto ya no está adjunta.');
      sub.attachmentIds=sub.attachmentIds.map(x=>x===oldId?id:x);if(sub.reportPhotoBoxes?.[oldId]){sub.reportPhotoBoxes[id]={...sub.reportPhotoBoxes[oldId],scale:next.reportScale,aspect:next.reportAspect};delete sub.reportPhotoBoxes[id].height;delete sub.reportPhotoBoxes[id].width;delete sub.reportPhotoBoxes[oldId];}
      // The new file is now referenced by the submission. A later UI error must
      // never delete it and leave a broken attachment behind.
      created=null;
      live.dataset.sglPhotoId=id;live.dataset.sglScale=String(next.reportScale);live.dataset.sglAspect=String(next.reportAspect);live.removeAttribute('data-sgl-height');const picture=live.querySelector('img');if(picture){picture.src=dataUrl;picture.dataset.sgl5048Ready=id;picture.dataset.sgl5047Ready=id;picture.style.visibility='visible';picture.loading='eager';}live.querySelector(':scope > .sgl5047-photo-status')?.remove();
      s.photo=id;changed(s);window.sgl5047MarkReportDirty?.();window.sgl512ApplyInlinePhotoLayout?.(live);window.sgl5048RenderRequirementFiles?.();window.sgl5039ScheduleReflow?.(180);
    }catch(e){if(created)try{await fileDelete(created);}catch{}toast('No se pudo girar la foto: '+e.message);}finally{s.busy=false;enhance();}
  }
  function enhance(){const s=state();if(!s)return;const modal=s.editor.closest('.modal'),active=desktop();modal.classList.toggle('sgl532-desktop',active);decoratePhotoTurns(s);if(!active)return;
    if(!s.initialized&&canEdit(s)){s.initialized=true;const sub=v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);if(canEdit(s)&&!sub.text&&!sub.richText){const font=document.getElementById('req_font'),size=document.getElementById('sgl5047_report_font_size');if(font){font.value='Arial';}if(size)size.value='12';s.editor.style.setProperty('--sgl5047-report-font-size','12pt');s.editor.dataset.sgl5047FontSize='12';}}
    if(!s.historyBound){s.historyBound=true;s.lastHtml=html();s.editor.addEventListener('input',()=>recordHistory(s));}
    let toolbar=modal.querySelector('.sgl5037-paper-toolbar');if(!toolbar){toolbar=document.createElement('div');toolbar.className='sgl5037-paper-toolbar';modal.querySelector('.sgl5037-paper-viewport')?.before(toolbar);}if(!toolbar.isConnected)return;if(!toolbar.querySelector('.sgl532-controls'))createBar(s,toolbar);
    const viewport=modal.querySelector('.sgl5037-paper-viewport'),shell=s.editor.closest('.sgl5031-page-shell');
    if(viewport&&shell&&s.viewportWidth!==viewport.clientWidth){s.viewportWidth=viewport.clientWidth;const fit=Math.min(1,Math.max(.35,(viewport.clientWidth-40)/(shell.offsetWidth||816)));window.sgl5037SetReportZoom?.(Math.floor(fit*100)/100);}
    for(const button of modal.querySelectorAll('.modal-foot button'))if((button.getAttribute('onclick')||'').includes('closeModal(')){button.classList.add('sgl532-footer-close');button.setAttribute('aria-label','Cerrar informe');button.title='Cerrar informe';}
    for(const input of toolbar.querySelectorAll('.sgl532-controls button:not([data-mic]):not([data-delete-report]),.sgl532-controls select,.sgl532-controls input'))input.disabled=!canEdit(s);
    const remove=modal.querySelector('[data-sgl5031-delete-progress-head],.sgl5088-delete-progress-modal'),trash=toolbar.querySelector('[data-delete-report]');
    if(trash){trash.hidden=!remove;trash.disabled=!remove||remove.disabled;}
    const originalMic=modal.querySelector('.sgl-leader-dictation-fab'),mic=toolbar.querySelector('[data-dictate]');
    mic.hidden=!originalMic;mic.disabled=!canEdit(s);
    if(originalMic&&s.micSource!==originalMic){
      s.micObserver?.disconnect();s.micSource=originalMic;
      const sync=()=>{const listening=originalMic.classList.contains('is-listening')||originalMic.getAttribute('aria-pressed')==='true';mic.setAttribute('aria-pressed',String(listening));mic.title=listening?'Detener dictado':'Dictar desde el cursor';mic.setAttribute('aria-label',mic.title);};
      s.micObserver=new MutationObserver(sync);s.micObserver.observe(originalMic,{attributes:true,attributeFilter:['class','aria-pressed']});sync();
    }
    for(const id of ['req_orientation','req_font'])document.getElementById(id)?.closest('label')?.classList.add('sgl532-legacy-field');
    modal.querySelectorAll('.sgl532-rotate').forEach(node=>node.remove());
    for(const button of toolbar.querySelectorAll('[data-history]'))button.disabled=!canEdit(s)||!(button.dataset.history==='undo'?s.history:s.redo).length;
    for(const button of toolbar.querySelectorAll('[data-turn]'))button.disabled=!canEdit(s)||s.busy||!selectedPhoto(s);

  }
  document.addEventListener('pointerdown',e=>{const s=state();if(s&&s.editor.contains(e.target))s.toolbarActive=false;},true);
  document.addEventListener('focusin',e=>{const s=state();if(s&&s.editor.contains(e.target))s.toolbarActive=false;},true);
  document.addEventListener('keydown',e=>{const s=state();if(s&&s.editor.contains(e.target))s.toolbarActive=false;},true);
  window.sgl605CaptureFormatSelection=()=>{const s=state();if(s)s.toolbarActive=false;capture(true);};
  document.addEventListener('selectionchange',capture);
  document.addEventListener('pointerup',e=>{const s=state(),photo=e.target.closest?.('#req_text figure[data-sgl-photo-id]');if(s&&photo){s.photo=photo.dataset.sglPhotoId;enhance();}capture();});
  document.addEventListener('keydown',e=>{
    const s=state();if(!canEdit(s)||!s.editor.contains(e.target)||!(e.ctrlKey||e.metaKey)||e.altKey)return;
    if(['b','i','u'].includes(e.key.toLowerCase())){e.preventDefault();capture();nativeCommand({b:'bold',i:'italic',u:'underline'}[e.key.toLowerCase()]);return;}
    const undo=e.key.toLowerCase()==='z'&&!e.shiftKey,redo=e.key.toLowerCase()==='y'||e.key.toLowerCase()==='z'&&e.shiftKey;if(!undo&&!redo)return;
    e.preventDefault();historyAction(s,undo?'undo':'redo');
  },true);
  let scheduled=false;const schedule=()=>{if(scheduled)return;scheduled=true;queueMicrotask(()=>{scheduled=false;enhance();});};
  const observer=new MutationObserver(schedule);observer.observe(document.getElementById('modal_root')||document.body,{childList:true,subtree:true});
  window.addEventListener('resize',schedule);window.sgl532EnhanceDesktop=enhance;window.sgl532RotatePhoto=rotate;schedule();
})();

/* 5.47 — shared report clipboard and document history, desktop and touch. */
(()=>{
  'use strict';
  const states=new WeakMap();let clip=null,busy=false;
  const key=()=>[currentUser()?.id,modalOwnerId,v3ModalProgramId,modalRequirementId].join('|');
  function current(){const editor=document.getElementById('req_text');if(!editor||!v3CanEditCurrentRequirementDraft?.())return null;let s=states.get(editor);if(!s){s={editor,key:key(),range:null,photos:new Set(),undo:[],redo:[],applying:false};states.set(editor,s);s.last=snapshot(s);editor.addEventListener('input',()=>record(s));}return s;}
  const valid=s=>s&&s.editor.isConnected&&s.key===key()&&v3CanEditCurrentRequirementDraft?.();
  const bodies=s=>[...s.editor.querySelectorAll('.sgl5039-page-body')];
  const sub=()=>v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);
  function snapshot(s){return {html:readRichEditor().html,ids:[...(sub().attachmentIds||[])],boxes:structuredClone(sub().reportPhotoBoxes||{})};}
  function equal(a,b){return JSON.stringify(a)===JSON.stringify(b);}
  function record(s){if(!valid(s)||s.applying)return;const next=snapshot(s);if(equal(next,s.last))return;s.undo.push(s.last);if(s.undo.length>60)s.undo.shift();s.redo=[];s.last=next;refresh(s);}
  function refresh(s){if(!valid(s))return;for(const b of s.editor.closest('.modal').querySelectorAll('[data-history],[data-action="undo"],[data-action="redo"]')){const action=b.dataset.history||b.dataset.action;b.disabled=!(action==='undo'?s.undo:s.redo).length;}}
  function changed(s){s.editor.closest('.modal').__sgl5047Dirty=true;s.editor.dispatchEvent(new Event('input',{bubbles:true}));window.sgl5047MarkReportDirty?.();window.sgl5039ScheduleReflow?.(0);}
  window.sgl539History=function(action){const s=current();if(!valid(s)||busy)return;const from=action==='undo'?s.undo:s.redo,to=action==='undo'?s.redo:s.undo;if(!from.length)return;to.push(snapshot(s));const value=from.pop();s.applying=true;sub().attachmentIds=[...value.ids];sub().reportPhotoBoxes=structuredClone(value.boxes);bodies(s).forEach((b,i)=>b.innerHTML=i?'':value.html);s.photos.clear();s.range=null;s.last=value;changed(s);s.applying=false;window.sgl5039ReflowNow?.(true);window.sgl512HydrateInlinePhotos?.().catch?.(()=>{});refresh(s);};
  function capture(s){const sel=getSelection();if(!sel?.rangeCount)return;const r=sel.getRangeAt(0),body=n=>(n.nodeType===1?n:n.parentElement)?.closest('.sgl5039-page-body');if(s.editor.contains(r.startContainer)&&s.editor.contains(r.endContainer)&&body(r.startContainer)&&body(r.endContainer))s.range=r.cloneRange();}
  function restore(s){let r=s.range;if(!r||!s.editor.contains(r.startContainer)||!s.editor.contains(r.endContainer)){toast('Coloca el cursor en la página donde deseas pegar.');return null;}const b=(r.startContainer.nodeType===1?r.startContainer:r.startContainer.parentElement).closest('.sgl5039-page-body');b?.focus({preventScroll:true});const sel=getSelection();sel.removeAllRanges();sel.addRange(r);window.__sgl5039ActiveBody=b;return r;}
  function parts(s,r){return bodies(s).filter(b=>r.intersectsNode(b)).map(b=>{const part=document.createRange();part.selectNodeContents(b);if(b.contains(r.startContainer))part.setStart(r.startContainer,r.startOffset);if(b.contains(r.endContainer))part.setEnd(r.endContainer,r.endOffset);return part;});}
  function selected(s){const box=document.createElement('div');const photos=[...s.editor.querySelectorAll('figure[data-sgl-photo-id]')].filter(f=>s.photos.has(f.dataset.sglPhotoId));if(photos.length){photos.forEach(f=>box.appendChild(f.cloneNode(true)));return {box,photos};}capture(s);const r=s.range;if(!r||r.collapsed)return null;for(const p of parts(s,r))box.appendChild(p.cloneContents());return {box,range:r.cloneRange()};}
  function clean(box){box.querySelectorAll('[data-sgl-editor-ui],.sgl5047-photo-status,.sgl5045-photo-quicktools').forEach(n=>n.remove());box.querySelectorAll('.sgl539-selected,.sgl5045-photo-selected').forEach(n=>n.classList.remove('sgl539-selected','sgl5045-photo-selected'));return box;}
  function prepare(s){const pick=selected(s);if(!pick){toast('Selecciona texto o una imagen primero.');return null;}clean(pick.box);const token=crypto.randomUUID();return {...pick,token,html:pick.box.innerHTML,text:pick.box.innerText||pick.box.textContent||'',owner:modalOwnerId,key:s.key,ids:[...new Set([...pick.box.querySelectorAll('figure[data-sgl-photo-id]')].map(f=>f.dataset.sglPhotoId))]};}
  function removeSelection(s,item){record(s);let caret=item.range?.cloneRange();if(item.photos?.[0]?.isConnected){caret=document.createRange();caret.setStartBefore(item.photos[0]);}caret?.collapse(true);const ids=item.ids;if(item.photos)s.editor.querySelectorAll('figure[data-sgl-photo-id]').forEach(f=>{if(ids.includes(f.dataset.sglPhotoId))f.remove();});else for(const r of parts(s,item.range).reverse())r.deleteContents();for(const b of bodies(s))if(!b.childNodes.length)b.innerHTML='<p><br></p>';const remaining=new Set([...s.editor.querySelectorAll('figure[data-sgl-photo-id]')].map(f=>f.dataset.sglPhotoId));sub().attachmentIds=(sub().attachmentIds||[]).filter(id=>!ids.includes(id)||remaining.has(id));for(const id of ids)if(!remaining.has(id)&&sub().reportPhotoBoxes)delete sub().reportPhotoBoxes[id];s.photos.clear();if(caret&&s.editor.contains(caret.startContainer)){s.range=caret;restore(s);}changed(s);}
  function htmlClip(item){return `<div data-sgl-clipboard-token="${item.token}">${item.html}</div>`;}
  async function copy(action,event){const s=current();if(!valid(s)||busy)return;const item=prepare(s);if(!item)return;clip=item;
    if(action==='cut')removeSelection(s,item);
    if(event?.clipboardData){event.preventDefault();event.stopImmediatePropagation();event.clipboardData.setData('text/html',htmlClip(item));event.clipboardData.setData('text/plain',item.text);event.clipboardData.setData('application/x-sgl-report',item.token);}
    else {try{if(navigator.clipboard?.write&&window.ClipboardItem)await navigator.clipboard.write([new ClipboardItem({'text/html':new Blob([htmlClip(item)],{type:'text/html'}),'text/plain':new Blob([item.text],{type:'text/plain'})})]);else throw Error('unavailable');}catch{toast('Copiado dentro de SGL. Usa el botón Pegar del informe.');}}

  }
  function internal(data){if(!clip)return false;const html=data?.getData?.('text/html')||'';return data?.getData?.('application/x-sgl-report')===clip.token||html.includes(`data-sgl-clipboard-token="${clip.token}"`);}
  function insertAt(r,fragment){const body=(r.startContainer.nodeType===1?r.startContainer:r.startContainer.parentElement).closest('.sgl5039-page-body');let block=r.startContainer.nodeType===1?r.startContainer:r.startContainer.parentElement;while(block&&block.parentElement!==body)block=block.parentElement;
    if(block&&block!==body&&/^(P|DIV|H[1-6])$/.test(block.tagName)&&fragment.querySelector('figure,p,div,ul,ol,table')){const tail=r.cloneRange();tail.setEnd(block,block.childNodes.length);const rest=tail.extractContents(),after=block.cloneNode(false);after.appendChild(rest);r.setStartAfter(block);r.collapse(true);if(after.textContent||after.querySelector('img,br'))block.after(after);}
    const last=fragment.lastChild;r.insertNode(fragment);if(last){r.setStartAfter(last);r.collapse(true);}getSelection().removeAllRanges();getSelection().addRange(r);
  }
  async function pasteItem(s,item){if(!valid(s)||busy)return;const r=restore(s);if(!r)return;if(item.owner&&item.owner!==modalOwnerId){toast('Pega las imágenes en un informe del mismo usuario.');return;}
    busy=true;const captured=r.cloneRange(),box=document.createElement('div');box.innerHTML=sanitizeRichText(item.html||'');const created=[];let committed=false;const initial=snapshot(s);
    try{
      const targetPage=Number((r.startContainer.nodeType===1?r.startContainer:r.startContainer.parentElement).closest('.sgl5039-word-page')?.dataset.sgl5039Page||1);
      for(const f of box.querySelectorAll('figure[data-sgl-photo-id]')){
        const source=await fileGet(f.dataset.sglPhotoId);if(!source?.dataUrl)throw Error('No se encontró una de las imágenes copiadas.');
        const id=newId('file'),file={...source,id,ownerId:modalOwnerId,programId:v3ModalProgramId||'class_grouped',requirementId:modalRequirementId,reportFree:false,reportPage:targetPage,reportPlacement:'auto',createdAt:isoNow()};for(const k of ['reportLeft','reportTop','reportHeight','reportZ'])delete file[k];await filePut(file);created.push(file);
        f.dataset.sglPhotoId=id;f.dataset.sglFree='0';f.dataset.sglPage=String(targetPage);for(const k of ['sglLeft','sglTop','sglZ','sglHeight'])delete f.dataset[k];f.removeAttribute('style');f.classList.remove('sgl539-selected','sgl5045-photo-selected');const img=f.querySelector('img');if(img){img.removeAttribute('src');delete img.dataset.sgl5048Ready;delete img.dataset.sgl5047Ready;}
      }
      if(!valid(s)||!s.editor.contains(captured.startContainer)||!s.editor.contains(captured.endContainer)||!equal(initial,snapshot(s)))throw Error('El informe cambió mientras se preparaba el pegado. Vuelve a colocar el cursor y pegar.');
      record(s);const deleted=[...s.editor.querySelectorAll('figure[data-sgl-photo-id]')].filter(f=>!captured.collapsed&&captured.intersectsNode(f)).map(f=>f.dataset.sglPhotoId);
      if(!captured.collapsed){for(const part of parts(s,captured).reverse())part.deleteContents();captured.collapse(true);}
      sub().attachmentIds=(sub().attachmentIds||[]).filter(id=>!deleted.includes(id));
      for(const f of created){sub().attachmentIds.push(f.id);(sub().reportPhotoBoxes||={})[f.id]={page:targetPage,free:false,scale:Number(f.reportScale||62)};}
      const frag=document.createDocumentFragment();frag.append(...box.childNodes);if(frag.lastElementChild?.tagName==='TABLE'){const p=document.createElement('p');p.innerHTML='<br>';frag.append(p);}insertAt(captured,frag);if(captured.startContainer.nodeType===1){const last=captured.startContainer.childNodes[captured.startOffset-1];if(last?.nodeType===1&&last.matches('p,div')&&!last.textContent){captured.selectNodeContents(last);captured.collapse(true);getSelection().removeAllRanges();getSelection().addRange(captured);}}committed=true;s.range=captured.cloneRange();s.photos.clear();changed(s);window.sgl5039ReflowNow?.(true);window.sgl606AfterPaste?.();capture(s);window.sgl512HydrateInlinePhotos?.().catch?.(()=>{});
    }catch(error){if(!committed)for(const f of created)try{await fileDelete(f.id);}catch{}toast('No se pudo pegar: '+error.message);}finally{busy=false;}
  }
  async function pasteButton(){const s=current();if(!valid(s))return;capture(s);try{
      if(navigator.clipboard?.read){const entries=await navigator.clipboard.read();for(const entry of entries){if(entry.types.includes('text/html')){const html=await (await entry.getType('text/html')).text();if(clip&&html.includes(clip.token))return pasteItem(s,clip);if(html)return pasteItem(s,{html:normalizeWordClipboardHtml(html)});}if(entry.types.some(t=>t.startsWith('image/'))){restore(s);const t=entry.types.find(t=>t.startsWith('image/'));return window.attachFiles?.([new File([await entry.getType(t)],'Imagen pegada',{type:t})]);}if(entry.types.includes('text/plain')){const text=await(await entry.getType('text/plain')).text();return pasteItem(s,{html:text.split('\n').map(l=>`<p>${esc(l)||'<br>'}</p>`).join('')});}}}
      else if(navigator.clipboard?.readText){const text=await navigator.clipboard.readText();return pasteItem(s,{html:text.split('\n').map(l=>`<p>${esc(l)||'<br>'}</p>`).join('')});}
    }catch{}
    if(clip?.key?.split('|')[0]===key().split('|')[0])return pasteItem(s,clip);toast('El navegador no permitió leer el portapapeles. Usa Ctrl+V o mantén pulsado el texto y elige Pegar.');
  }
  // Window capture runs before legacy document paste handlers, keeping internal
  // photo references out of the external-image import pipeline.
  for(const action of ['copy','cut'])window.addEventListener(action,e=>{const s=current();if(valid(s)&&(s.editor.contains(e.target)||s.photos.size))copy(action,e);},true);
  window.addEventListener('paste',e=>{const s=current();if(!valid(s)||!s.editor.contains(e.target)||!internal(e.clipboardData))return;e.preventDefault();e.stopImmediatePropagation();capture(s);pasteItem(s,clip);},true);
  window.addEventListener('keydown',e=>{const s=current();if(!valid(s)||!(e.ctrlKey||e.metaKey)||e.altKey||!s.editor.contains(e.target))return;const k=e.key.toLowerCase();if((k==='c'||k==='x')&&s.photos.size){e.preventDefault();e.stopImmediatePropagation();copy(k==='x'?'cut':'copy');return;}if(k==='z'||k==='y'){e.preventDefault();e.stopImmediatePropagation();window.sgl539History(k==='y'||e.shiftKey?'redo':'undo');}},true);
  window.addEventListener('click',e=>{const b=e.target.closest?.('[data-history],[data-action="undo"],[data-action="redo"]');if(!b||!current()||!b.closest('.modal'))return;e.preventDefault();e.stopImmediatePropagation();window.sgl539History(b.dataset.history||b.dataset.action);},true);
  document.addEventListener('selectionchange',()=>{const s=current();if(valid(s))capture(s);});
  window.addEventListener('pointerup',e=>{const s=current();if(!valid(s)||!s.editor.contains(e.target))return;const f=e.target.closest('figure[data-sgl-photo-id]');if(f&&!e.target.closest('button')){if(!e.ctrlKey&&!e.metaKey&&!e.shiftKey&&window.matchMedia('(min-width:821px)').matches)s.photos.clear();if(s.photos.has(f.dataset.sglPhotoId))s.photos.delete(f.dataset.sglPhotoId);else s.photos.add(f.dataset.sglPhotoId);}else if(!f)s.photos.clear();for(const figure of s.editor.querySelectorAll('figure[data-sgl-photo-id]'))figure.classList.toggle('sgl539-selected',s.photos.has(figure.dataset.sglPhotoId));capture(s);},true);
  function enhance(){const s=current();if(!valid(s))return;const modal=s.editor.closest('.modal'),host=modal.querySelector('.sgl537-page-actions');if(host&&!host.querySelector('.sgl539-clipboard')){const bar=document.createElement('div');bar.className='sgl539-clipboard';bar.setAttribute('role','group');bar.setAttribute('aria-label','Portapapeles del informe');bar.innerHTML=['cut','copy','paste'].map((a,i)=>`<button type="button" data-clipboard="${a}" aria-label="${['Cortar','Copiar','Pegar'][i]}" title="${['Cortar (Ctrl+X)','Copiar (Ctrl+C)','Pegar (Ctrl+V)'][i]}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${['<circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="m8.5 7.5 12 12M8.5 16.5 20 4M12 12l3-3"/>','<rect x="9" y="8" width="11" height="13" rx="2"/><path d="M15 8V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h4"/>','<path d="M9 5H6a2 2 0 0 0-2 2v13h16V7a2 2 0 0 0-2-2h-3"/><rect x="9" y="3" width="6" height="4" rx="1"/><path d="M8 12h8m-8 4h6"/>'][i]}</svg></button>`).join('');bar.addEventListener('pointerdown',e=>{capture(s);e.preventDefault();});bar.addEventListener('click',e=>{const action=e.target.closest('button')?.dataset.clipboard;if(action==='paste')pasteButton();else if(action)copy(action);});host.prepend(bar);}refresh(s);}
  let scheduled=false;const observer=new MutationObserver(()=>{if(scheduled)return;scheduled=true;setTimeout(()=>{scheduled=false;enhance();},0);});observer.observe(document.getElementById('modal_root')||document.body,{childList:true,subtree:true});window.sgl539Clipboard={copy,paste:pasteButton,pasteItem,current};enhance();
})();
