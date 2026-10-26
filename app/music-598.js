/* Música 6.0: explicit scope, opt-in menus and manual contact. */
'use strict';
(()=>{
 const P='music_ministry',music=()=>!!window.__MUSICA_WEB__||window.SGL562?.view()==='mmus',actual=()=>window.SGL555?.actual()||currentUser();
 const prefs=()=>database.music550?.interface598||{},drive=()=>false;
 const personal=()=>!!window.SGL555?.personalMode()||['aspirant','pastor_aspirant'].includes(currentUser()?.role);
 const eligible=u=>SGLMusicPolicy560.active(u)&&['admin','departamental','secretaria','coordinador_musica','administrativo','regional'].includes(SGLMusicPolicy560.role?.(u)||({administrativo_sgm:'secretaria'})[u?.role]||u?.role)&&SGL555.managementPrograms(u).includes(P);
 const noTutor=u=>!(Object.hasOwn(u,'tutorByProgram')?u.tutorByProgram?.[P]:u.tutorId);
 const nav=navItems;navItems=function(u){let n=nav.apply(this,arguments);if(!music())return n;n=n.filter(([id])=>!['music_unassigned','bible'].includes(id)&&(id!=='drive'||drive()));return n.map(([id,icon,label,...rest])=>[id,icon,!personal()&&['daily_activity','requirements','music_module'].includes(id)&&/Mi ruta|Trayectoria/i.test(label)?'Avances':label,...rest]);};window.navItems=navItems;
 const navigate=go;go=function(page,...args){if(music()){if(page==='music_unassigned')page='aspirants';if(page==='drive')return;if(page==='bible')return;}return navigate.call(this,page,...args);};window.go=go;
 const oldDrive=renderDrive;renderDrive=function(){if(music()&&!drive())return go('dashboard');return oldDrive.apply(this,arguments);};window.renderDrive=renderDrive;
 const contactAllowed=()=>false;const contactButton=()=>'';
 const driveEnabled=program=>program!==P||drive();
 function reviewLabel(owner,status,program){return status==='in_review'&&(program!==P||!personal()&&v3CanReviewRequirement(currentUser(),owner,status,program))?'Revisar':'Ver';}
 function decorate(){
  if(!music()){document.title='Sistema de Gestión de Liderazgo';const brand=document.querySelector('.side-brand strong');if(brand)brand.textContent='Sistema de Gestión de Liderazgo';return;}const root=document.getElementById('content');if(!root)return;
  if(['dashboard','music_module'].includes(currentPage))for(const el of root.querySelectorAll('button,a'))if(/Lectura B[ií]blica/i.test(el.textContent))el.remove();
  if(!personal()){for(const el of root.querySelectorAll('h1,h2,h3,button,a'))if(/^(Mi ruta|Trayectoria)( de Música)?$/.test(el.textContent.trim()))el.textContent='Avances';}
  compactHeadings(root);
  if(currentPage==='aspirants'&&!personal()&&!document.getElementById('sgm598_unassigned')){
   const people=accessibleAspirants().filter(noTutor),section=document.createElement('section');section.id='sgm598_unassigned';section.className='sgm601-unassigned';section.setAttribute('aria-label','Aspirantes sin tutor');section.innerHTML=`<strong class="sgm601-unassigned-title">Sin tutor (${people.length})</strong><div class="sgm601-unassigned-list">`+(people.map(u=>`<article class="sgm601-unassigned-person"><strong>${esc(u.fullName)}</strong><small>${esc(u.region||'Sin región')}</small><button class="btn ghost small" onclick="${esc('SGL561.assign('+JSON.stringify(u.id)+')')}">Asignar tutor</button>${contactButton(u)}</article>`).join('')||'<span>No hay aspirantes sin tutor en tu alcance.</span>')+'</div>';root.querySelector('.page-head')?.after(section);
  }
  syncContacts();
  if(currentPage==='settings'&&actual()?.role==='admin'&&!personal()&&!document.getElementById('sgm598_options')){
   const card=document.createElement('section');card.id='sgm598_options';card.className='card';const ids=prefs().whatsappUsers||[];
   card.innerHTML='';
  }
 }
 const content=setContent;setContent=function(...args){const result=content.apply(this,args);decorate();return result;};window.setContent=setContent;
 // Keep the old link callable, but render its list inside Aspirantes.
 SGL561.unassigned=()=>go('aspirants');
 const phrases=['¡Excelente trabajo! Has presentado un nuevo informe.','¡Felicitaciones por tu constancia! Tu dedicación se refleja en esta entrega.','¡Un paso más en tu formación! Gracias por completar este informe.','¡Muy bien! Cada entrega fortalece tu preparación para servir.','¡Sigue creciendo! Hoy has sumado un nuevo esfuerzo a tu formación.','¡Tu perseverancia cuenta! Has completado otra entrega importante.','¡Gracias por tu compromiso! Tu informe ya forma parte de tu recorrido.','¡Celebra este logro! Tu esfuerzo merece reconocimiento.','¡Adelante con confianza! Tu disciplina está dando frutos.','¡Buen trabajo! Una nueva entrega demuestra tu responsabilidad.','¡Tu formación sigue avanzando! Reconocemos el empeño de esta entrega.','¡Enhorabuena! Has dado otro paso con propósito y dedicación.','¡Bien hecho! Tu preparación se construye con esfuerzos como este.','¡Qué buena entrega! Continúa cultivando tus dones con entusiasmo.','¡Tu compromiso deja huella! Gracias por perseverar en tu aprendizaje.','¡Cada paso tiene valor! Hoy has cumplido una nueva meta de trabajo.','¡Felicidades por esta entrega! Tu disposición para aprender es valiosa.','¡Una nueva meta presentada! Que este logro te anime a seguir formándote.','¡Vas construyendo tu camino! Tu informe refleja tiempo y dedicación.','¡Celebra tu perseverancia! Has mantenido el compromiso con tu preparación.','¡Un esfuerzo que merece reconocimiento! Gracias por compartir tu trabajo.','¡Suma este logro a tu recorrido! Tu formación crece con cada entrega.','¡Qué alegría recibir tu trabajo! Sigue desarrollando tus talentos.','¡Tu disciplina se hace visible! Has presentado otro informe de tu formación.','¡Un nuevo motivo para seguir! Reconocemos tu interés por aprender.','¡Lo has presentado! Gracias por dar continuidad a tu proceso formativo.','¡Tu esfuerzo de hoy cuenta! Sigue preparando tus dones para el servicio.','¡Muy buena constancia! Esta entrega es parte de tu crecimiento.','¡Has cumplido otra etapa de trabajo! Valora lo que estás aprendiendo.','¡Tu dedicación inspira! Gracias por avanzar con responsabilidad.','¡Un logro más para celebrar! Sigue formándote con esperanza y propósito.'];
 function congratulation(owner,requirement,token,announce=false){
  const key='sgm598_congrats_'+owner;let state={count:0,seen:{},announced:{}};try{state=JSON.parse(localStorage.getItem(key))||state;}catch{}state.seen=state.seen||{};state.announced=state.announced||{};
  const identity=String(token||requirement);if(!Object.hasOwn(state.seen,identity)){state.seen[identity]=(Number(state.count)||0)%phrases.length;state.count=(Number(state.count)||0)+1;}
  const message=phrases[state.seen[identity]]+' Informe recibido para revisión; todavía no está aprobado.';
  const show=announce&&!state.announced[identity]&&window.SGL602?.permit('content.showNotices')!==false;if(show)state.announced[identity]=true;
  try{localStorage.setItem(key,JSON.stringify(state));}catch{}
  if(show)setTimeout(()=>{if(actual()?.id===owner&&music())toast(message);},100);
  return message;
 }
 const reviewMessage=sgl538ReviewMessage;sgl538ReviewMessage=function(owner=modalOwnerId,program=v3ModalProgramId||'class_grouped',requirement=modalRequirementId){
  const base=reviewMessage.apply(this,[owner,program,requirement]);if(program!==P||owner!==actual()?.id||!['aspirant','pastor_aspirant'].includes(currentUser()?.role))return base;
  const match=op=>op.action==='review'&&op.ownerId===owner&&op.programId===P&&op.requirementId===requirement;
  const receipt=(database.syncHistory||[]).find(match);if((database.syncQueue||[]).some(match)||!receipt?.filesReceived)return base;
  return congratulation(owner,requirement,receipt.id||receipt.operationId||requirement);
 };window.sgl538ReviewMessage=sgl538ReviewMessage;
 const api=v3Api;v3Api=async function(path,options,...args){const result=await api.call(this,path,options,...args),body=options?.body,payload=result?.payload;
  if(String(path).split('?')[0].endsWith('/api/reviews/submit')&&body?.programId===P&&body.ownerId===actual()?.id&&['aspirant','pastor_aspirant'].includes(currentUser()?.role)&&payload?.filesReceived===true&&payload.currentStatus==='in_review')congratulation(body.ownerId,body.requirementId,body.operationId,true);
  return result;
 };window.v3Api=v3Api;
 function compactHeadings(root){
  // Only page headers: repeated field labels and section headings remain intact.
  for(const header of root.querySelectorAll('.page-head')){
   const title=header.querySelector('h1');if(!title)continue;
   for(const eyebrow of header.querySelectorAll('.eyebrow'))if(eyebrow.textContent.trim()===title.textContent.trim()||/^Sistema de Gestión de Música$/i.test(eyebrow.textContent.trim()))eyebrow.remove();
   for(const subtitle of header.querySelectorAll('.subtitle'))if(subtitle.textContent.trim()===title.textContent.trim()||subtitle.textContent.trim()==='Personas y expedientes dentro de tu alcance.')subtitle.remove();
  }
 }
 let pending598=false;function scheduleControls598(){if(pending598)return;pending598=true;queueMicrotask(()=>{pending598=false;syncContacts();});}
 function syncContacts(){
  if(!music())return;const root=document.getElementById('content');if(!root)return;
  for(const area of [root,document.getElementById('modal_root')].filter(Boolean)){
   for(const label of area.querySelectorAll('.page-head h1,.page-head .eyebrow,.chip,.sgl551-owner-name')){
    for(const node of label.childNodes)if(node.nodeType===3&&node.textContent.includes('Sistema de Gestión de Música SGM'))node.textContent=node.textContent.replaceAll('Sistema de Gestión de Música SGM','Sistema de Gestión de Música');
   }
   if(!drive()){
    for(const option of area.querySelectorAll('option[value="drive"]')){const select=option.parentElement,selected=option.selected;option.remove();if(selected&&select){select.value=select.querySelector('option[value="all"]')?'all':'both';select.dispatchEvent(new Event('change',{bubbles:true}));}}
    for(const control of area.querySelectorAll('#sgl525_verify_drive_all,.sgl424-drive-cell,button[onclick*="sgl424CheckDriveFile"],button[onclick*="sgl424OpenDriveLink"],button[onclick*="sgl424CopyDriveLink"],button[onclick*="sgl424EnableDriveReaderLink"],button[onclick*="sgl424DeleteDriveFile"],button[onclick*="sendReportToDrive"]'))control.remove();
   }
  }

 }
 // Report rows and action cells can be replaced independently of setContent.
 const host=document.getElementById('content');if(host)new MutationObserver(scheduleControls598).observe(host,{childList:true,subtree:true});
 const modalHost=document.getElementById('modal_root');if(modalHost)new MutationObserver(scheduleControls598).observe(modalHost,{childList:true,subtree:true});
 if(typeof v3ApplyServerSnapshot==='function'){const apply=v3ApplyServerSnapshot;v3ApplyServerSnapshot=async function(...args){const r=await apply.apply(this,args);syncContacts();return r;};window.v3ApplyServerSnapshot=v3ApplyServerSnapshot;}
 const style=document.createElement('style');style.textContent='.sgm601-contact{display:inline-flex;align-items:center;gap:4px;flex-wrap:nowrap}.sgm601-unassigned{display:flex;align-items:center;gap:14px;padding:10px 14px;margin:0 0 12px;border:1px solid #dce2ee;border-radius:12px;min-width:0}.sgm601-unassigned-title{flex:none;white-space:nowrap}.sgm601-unassigned-list{display:flex;align-items:center;gap:12px;overflow-x:auto;min-width:0;flex:1;padding:2px 0}.sgm601-unassigned-person{display:inline-flex;align-items:center;gap:10px;white-space:nowrap;flex:none;padding:0 12px 0 0;border-right:1px solid #dce2ee}.sgm601-unassigned-person:last-child{border-right:0}.sgm601-unassigned-person strong{font-size:13px}.sgm601-unassigned-person .btn{padding:7px 10px;min-height:32px}.sgm601-unassigned-person small{opacity:.75}';document.head.append(style);
 window.SGL598={decorate,contactAllowed,contactMarkup:contactButton,driveEnabled,reviewLabel};
})();
