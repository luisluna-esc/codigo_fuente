'use strict';

function sgl556ItemProgram(item){
 if(item?.programId)return item.programId;
 const entity=String(item?.requirementId||item?.entity||'');
 for(const [prefix,p] of [['LM-','music_ministry'],['GMMA-','guide_master_advanced'],['GMM-','guide_master'],['GM-','guide_major']])if(entity.startsWith(prefix))return p;
 const patch=item?.patch||{},ids=new Set();
 for(const programs of Object.values(patch.programProgress||{}))for(const id of Object.keys(programs||{}))ids.add(id);
 for(const key of ['tasks','materials','indicationRules'])for(const x of Array.isArray(patch[key])?patch[key]:[])ids.add(x.programId||'class_grouped');
 return ids.size===1?[...ids][0]:'class_grouped';
}

(()=>{
 const MUSIC='music_ministry',creatorRoles=['departamental','administrativo','secretaria','administrativo_sgm'];
 const activeProgram=()=>window.__MUSICA_WEB__?MUSIC:(v3ActiveProgramId||'class_grouped');
 const actual=()=>window.SGL555?.actual()||currentUser();
 const isManagement=()=>!window.SGL555?.personalMode();
 const canCreate=()=>isManagement()&&(activeProgram()===MUSIC?SGLMusicPolicy560.cap(currentUser(),'createTutor'):creatorRoles.includes(currentUser()?.role))&&window.SGL555?.managementPrograms(actual()).includes(activeProgram());
 const allowedPrograms=()=>activeProgram()===MUSIC?[MUSIC]:(window.SGL555?.managementPrograms(actual())||[]);
 const api=v3Api;
 v3Api=function(path,options={}){
  const program=options.body?.programId||(options.body?.patch?sgl556ItemProgram({patch:options.body.patch}):activeProgram());
  return api.call(this,path,{...options,headers:{...options.headers,'X-SGL-Program':program}});
 };window.v3Api=v3Api;
 const audit=addAudit;
 addAudit=function(...args){const result=audit.apply(this,args);const row=database.audit?.[0];if(row)row.programId=/^(LM|GMMA|GMM|GM)-/.test(String(args[1]||''))?sgl556ItemProgram({entity:args[1]}):activeProgram();return result;};window.addAudit=addAudit;
 const apply=v3ApplyServerSnapshot;
 v3ApplyServerSnapshot=async function(snapshot,...args){
  const tag=snapshot?.__accountView555;
  if(tag&&(tag.actor!==actual()?.id||tag.view!==(window.SGL555?.personalMode()?'personal':'management')))return {stale:true};
  const result=await apply.call(this,snapshot,...args);enhance();return result;
 };window.v3ApplyServerSnapshot=v3ApplyServerSnapshot;
 const review=v3CanReviewRequirement;
 v3CanReviewRequirement=function(actor,owner,status='pending'){
  if(actor?.id===owner?.id)return false;
  const p=typeof v3ModalProgramId==='string'&&document.querySelector('#modal_root')?v3ModalProgramId:v3ActiveProgramId;
  if(p===MUSIC&&status==='in_review'&&['regional',...creatorRoles,'coordinador_musica'].includes(actor?.role))return !!window.SGL555?.canSee(actor,owner,MUSIC);
  return review.apply(this,arguments);
 };window.v3CanReviewRequirement=v3CanReviewRequirement;
 const relevant=sgl529NoticeRelevant;
 sgl529NoticeRelevant=function(item,user){const program=item?.programId||(String(item?.requirementId||'').startsWith('LM-')?MUSIC:/^admin_change/.test(item?.type||'')?'class_grouped':'');return (!program||program===activeProgram())&&relevant.call(this,item,user);};window.sgl529NoticeRelevant=sgl529NoticeRelevant;
 const nav=navItems;
 navItems=function(actor){
  let items=nav.call(this,actor);
  if(activeProgram()===MUSIC){items=items.filter(x=>x[0]!=='materials');if(['aspirant','pastor_aspirant'].includes(actor?.role)&&!items.some(x=>x[0]==='requirements'))items.push(['requirements','▤','Mi Tarjeta Líder']);const i=items.findIndex(x=>x[0]==='requirements');if(i>=0){const item=[...items.splice(i,1)[0]];if(['aspirant','pastor_aspirant'].includes(actor?.role))item[2]='Mi Tarjeta Líder';items.splice(Math.max(0,items.findIndex(x=>x[0]==='dashboard')+1),0,item);}}
  if(canCreate()&&!items.some(x=>x[0]==='tutors'))items.splice(Math.max(0,items.findIndex(x=>x[0]==='aspirants')+1),0,['tutors','♙','Tutores']);
  return items;
 };window.navItems=navItems;
 const tutors=renderTutors;
 renderTutors=function(...args){
  if(!canCreate())return tutors.apply(this,args);
  const program=activeProgram(),people=(database.users||[]).filter(u=>u.role==='tutor'&&sgl421AccountActive(u)&&sgl548TutorPrograms(u).includes(program)).sort((a,b)=>String(a.fullName||'').localeCompare(String(b.fullName||''),'es'));
  setTop('Tutores',ROLE_LABELS[currentUser()?.role]||'Gestión');
  setContent(`<div class="page-head"><div><h1 class="page-title">Tutores</h1><p class="subtitle">Tutores del módulo ${esc(v3Program(program).menu_label||v3Program(program).name)}.</p></div><div class="actions"></div></div><div class="sgl5060-tutor-list">${people.map(u=>`<article class="card sgl5060-tutor-row"><strong>${esc(u.fullName||u.username||'Tutor')}</strong><button class="btn ghost small" onclick="sgl5060ShowTutorAspirants('${esc(u.id)}')">Ver Aspirantes asignados</button></article>`).join('')||'<div class="card empty">Todavía no hay tutores en este módulo.</div>'}</div>`);
 };window.renderTutors=renderTutors;
 const content=setContent;
 setContent=function(...args){const result=content.apply(this,args);enhance();return result;};window.setContent=setContent;
 const modal=showModal;
 showModal=function(...args){const result=modal.apply(this,args);enhanceModal();return result;};window.showModal=showModal;
 const filtered=filterRequirements;
 filterRequirements=function(...args){const result=filtered.apply(this,args);decorateWorkflow();return result;};window.filterRequirements=filterRequirements;

 function regionSelect(input){
  if(input.tagName!=='INPUT')return;
  const select=document.createElement('select');for(const a of input.attributes)if(!['type','placeholder','value','min','max'].includes(a.name))select.setAttribute(a.name,a.value);
  const music=activeProgram()===MUSIC;const selectedMusic=music?SGLMusicPolicy560.region(input.value):'';if(music)select.required=true;const region=String(input.value||'').match(/^(?:regi[oó]n\s*)?(\d{1,2})$/i)?.[1]||'';
  select.innerHTML='<option value="">Seleccionar región</option>'+Array.from({length:12},(_,i)=>`<option value="${i+1}">${i+1}</option>`).join('')+(music?'<option value="MBOS">MBOS</option>':'');select.value=music?selectedMusic:String(Number(region)||'');if(input.readOnly)select.disabled=true;input.replaceWith(select);
 }
 function enhanceModal(){
  for(const input of document.querySelectorAll('input[id*="region"],input[name="region"],input[data-profile548="region"],input[data-registration548="region"]'))regionSelect(input);
  const root=document.querySelector('#modal_root');if(!root)return;
  const seen=new Set();for(const b of root.querySelectorAll('button')){const call=b.getAttribute('onclick')||'';if(!/reviewRequirement\(['"]approved['"]\)/.test(call)&&!/^Marcar revisado$/i.test((b.textContent||'').trim()))continue;const key='approve';if(seen.has(key))b.remove();else seen.add(key);}
  // Music authority accounts may have their institutional reviewer assigned.
  if(document.getElementById('user_role')&&document.getElementById('sgl545_tutors')){
   const rg=document.getElementById('sgl545_tutors');if(document.querySelector('[name=user_module][value="music_ministry"]:checked'))rg.hidden=false;
   const field=rg.querySelector('[data-sgl545-program="music_ministry"]');
   if(field){for(const option of [...field.options])if(option.value&&option.value===rg.dataset.ownerId)option.remove();}
   if(field)for(const u of database.users||[])if(u.id!==rg.dataset.ownerId&&['admin','secretaria','administrativo','departamental','regional','coordinador_musica','administrativo_sgm'].includes(u.role)&&sgl421AccountActive(u)&&window.SGL555?.managementPrograms(u).includes(MUSIC)&&![...field.options].some(o=>o.value===u.id))field.add(new Option(u.fullName,u.id));
  }
 }
 function enhance(){
  enhanceModal();const host=document.getElementById('content');if(!host)return;
  if((currentPage==='dashboard'||currentPage==='music_module'&&SGL550.getSection()==='home')&&activeProgram()===MUSIC&&!document.getElementById('sgl556_materials')){
   const b=document.createElement('button');b.id='sgl556_materials';b.className='btn ghost';b.textContent='Materiales de apoyo';b.onclick=()=>go('materials');(host.querySelector('.page-head .actions')||host.querySelector('.page-head')||host).append(b);
  }
  if(canCreate()&&currentPage==='tutors'&&!document.getElementById('sgl556_create_tutor')){
   const b=document.createElement('button');b.id='sgl556_create_tutor';b.className='btn primary';b.textContent='Crear tutor';b.onclick=openTutor;(host.querySelector('.page-head .actions')||host.querySelector('.page-head')||host).append(b);
  }
  decorateWorkflow();
 }
 let changing=false;
 function decorateWorkflow(){
  const host=document.getElementById('content');if(!host||currentPage!=='requirements'||v3ActiveProgramId!==MUSIC)return;
  const owner=targetAspirant?.();if(!owner)return;const flow=database.music550?.workflow?.[owner.id]?.state||'',reqs=v3ProgramRequirements(MUSIC),p=v3EnsureProgramProgress(owner.id,MUSIC),ready=reqs.length>0&&reqs.every(r=>p.status?.[r.id]==='approved');
  const label=flow==='approved'?'Aprobado':flow==='recommended'?'Recomendado':'';
  host.dataset.musicWorkflow=label?flow:'';
  for(const row of host.querySelectorAll('[data-sgl5059-requirement],[data-sgl5058-requirement]')){row.classList.toggle('sgl556-recommended',flow==='recommended');row.classList.toggle('sgl556-approved',flow==='approved');const line=row.querySelector('.sgl5059-status-line');if(line&&label){const html=`<span class="status ${flow}">${esc(label)}</span>`;if(line.innerHTML!==html)line.innerHTML=html;}}
  let panel=document.getElementById('sgl556_workflow');if(!panel){panel=document.createElement('div');panel.id='sgl556_workflow';panel.className='actions';(host.querySelector('.page-head')||host).append(panel);}
  const role=currentUser()?.role,self=currentUser()?.id===owner.id;
  const action=ready&&!self&&flow!=='approved'?(flow==='recommended'&&SGLMusicPolicy560.final(currentUser(),owner,v3ModuleRecord(MUSIC)?.assignedUserIds||[])?'approve':flow!=='recommended'&&SGLMusicPolicy560.cap(currentUser(),'recommend')?'recommend':''):'';
  const html=(label?`<span class="chip">${esc(label)}</span>`:'')+(action?`<button type="button" class="btn primary" data-workflow556="${action}" ${changing?'disabled':''}>${action==='approve'?'APROBADO':'Recomendar'}</button>`:'');
  if(panel.innerHTML!==html)panel.innerHTML=html;
  const b=panel.querySelector('[data-workflow556]');if(b)b.onclick=()=>workflowAction(action,owner.id);
 }
 async function workflowAction(action,id){if(changing)return;changing=true;decorateWorkflow();try{await SGL550.action(action,id);decorateWorkflow();}finally{changing=false;decorateWorkflow();}}
 function openTutor(){
  if(!canCreate())return;const programs=allowedPrograms().filter(id=>v3ModuleRecord(id)?.active!==false);
  showModal(`<div class="modal small"><div class="modal-head"><h3>Crear tutor</h3><p>Después de crear un tutor debes asignar un Aspirante.</p><button class="modal-close" onclick="closeModal()">×</button></div><form id="sgl556_tutor_form"><div class="modal-body"><label class="field-group">Nombres y Apellidos Completos<input class="field" name="fullName" required maxlength="200"></label><label class="field-group">Correo<input class="field" name="email" type="email" required></label><label class="field-group">Usuario<input class="field" name="username" required maxlength="100"></label><label class="field-group">Región<select class="field" name="region" ${activeProgram()===MUSIC?'required':''}><option value="">Seleccionar región</option>${Array.from({length:12},(_,i)=>`<option>${i+1}</option>`).join('')}${activeProgram()===MUSIC?'<option value="MBOS">MBOS</option>':''}</select></label><label class="field-group">Iglesia<input class="field" name="church"></label><label class="field-group">Teléfono<input class="field" name="phone" type="tel"></label><label class="field-group">Contraseña temporal<input class="field" name="password" type="password" minlength="8" required autocomplete="new-password"></label><fieldset><legend>Módulos de tutoría</legend>${programs.map(p=>`<label class="choice"><input type="checkbox" name="program" value="${esc(p)}" ${p===activeProgram()?'checked':''}>${esc(v3Program(p).menu_label||v3Program(p).name)}</label>`).join('')}</fieldset><p class="error" id="sgl556_tutor_error"></p></div><div class="modal-foot"><button class="btn primary">Crear tutor</button></div></form></div>`);
  const form=document.getElementById('sgl556_tutor_form'),id=newId('usr');if(activeProgram()===MUSIC&&currentUser()?.role==='regional'){form.elements.region.value=SGLMusicPolicy560.region(currentUser().region);form.elements.region.disabled=true;}
  form.onsubmit=async event=>{event.preventDefault();const b=form.querySelector('.modal-foot button');b.disabled=true;try{const values=new FormData(form),programIds=values.getAll('program');if(!programIds.length)throw new Error('Selecciona al menos un módulo.');const user=Object.fromEntries(values);if(activeProgram()===MUSIC&&currentUser()?.role==='regional')user.region=SGLMusicPolicy560.region(currentUser().region);user.password=await passwordRecord(values.get('password'));delete user.program;await v3Api('/api/accounts/tutor',{method:'POST',body:{actorId:currentUser().id,id,programId:activeProgram(),programIds,restoreGeneration:database.restoreGeneration||1,user},timeout:20000});rosterHash='';await refreshRoster();await closeModal();toast('Tutor creado. Después de crear un tutor debes asignar un Aspirante.');}catch(error){setText('sgl556_tutor_error',error.message);}finally{b.disabled=false;}};
 }
 let rosterHash='',rosterActor='',polling=false,lastRosterAt=0;
 async function refreshRoster(){
  const actor=currentUser();if(polling||!actor||document.hidden||!database.settings?.serverUrl||!v3CanAttemptNetwork())return;
  if(document.querySelector('#modal_root [contenteditable="true"],#modal_root textarea')||window.__sgl555Switching)return;
  polling=true;const key=[actor.id,actor.role,window.SGL555?.personalMode(),window.__MUSICA_WEB__].join('|');if(key!==rosterActor){rosterActor=key;rosterHash='';}
  if(rosterHash&&Date.now()-lastRosterAt<15000){polling=false;return;}lastRosterAt=Date.now();
  try{const {payload}=await v3Api('/api/accounts/roster?'+new URLSearchParams({actorId:actor.id,knownHash:rosterHash}),{timeout:12000});if(currentUser()?.id!==actor.id)return;if(payload.database){const result=await v3ApplyServerSnapshot(payload);if(result===false||result?.stale||result?.deferred)return;if(currentPage==='requirements')filterRequirements();}rosterHash=payload.rosterHash||'';}catch{}finally{polling=false;}
 }
 window.SGL556={refreshRoster,openTutor,enhance,workflowAction,decorateWorkflow};
 setInterval(refreshRoster,15000);window.addEventListener('focus',refreshRoster);window.addEventListener('online',refreshRoster);
 const openForty=SGL531.open;
 SGL531.open=async function(req,owner,force){
  if(req!=='LM-TEM-1'||force==='readings')return openForty.apply(this,arguments);
  owner=owner||targetAspirant()?.id||currentUser()?.id;const ctx=SGL531.context(req,owner);
  if(!ctx.editable)return openForty.call(this,req,owner,force);
  if(!await SGL531.beforeLeave())return;
  v3SetReportContext(MUSIC,owner,req);
  showModal(`<div class="modal small"><div class="modal-head"><h3>40 días / 40 madrugadas</h3><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><p>Elige cómo presentar tu participación.</p><div class="actions"><button class="btn primary" id="sgl556_forty_pdf">Subir certificado PDF</button><button class="btn ghost" id="sgl556_forty_readings">Continuar llenando lecturas</button></div><p>Ambas modalidades se envían a revisión. Tus registros anteriores se conservan.</p></div></div>`);
  document.getElementById('sgl556_forty_pdf').onclick=()=>{v3SetReportContext(MUSIC,owner,req);window.sgl531OpenDocumentRequirement(req);};
  document.getElementById('sgl556_forty_readings').onclick=()=>openForty.call(SGL531,req,owner,'readings');
 };
 const openReq=openRequirement;
 openRequirement=function(req,...args){if(v3ActiveProgramId===MUSIC&&req==='LM-TEM-1')return SGL531.open(req,modalOwnerId||targetAspirant()?.id);return openReq.call(this,req,...args);};window.openRequirement=openRequirement;
 const style=document.createElement('style');style.textContent='.sgl556-recommended{background:#f1e8ff!important;border-color:#a66be2!important}.sgl556-approved{background:#e3f5eb!important;border-color:#388c5e!important}[data-music-workflow="recommended"] .status.recommended{background:#e8d7ff;color:#512078}[data-music-workflow="approved"] .status.approved{background:#d2efdd;color:#165d35}html[data-sgl545-mode="dark"] .sgl556-recommended{background:#38224c!important}html[data-sgl545-mode="dark"] .sgl556-approved{background:#173c2b!important}.sgl-journey-badge.unstarted{background:#eef0f4;color:#525d6b}#sgl556_workflow{flex-wrap:wrap}';document.head.append(style);
 let scheduled=false;new MutationObserver(()=>{if(scheduled)return;scheduled=true;setTimeout(()=>{scheduled=false;enhanceModal();},0);}).observe(document.getElementById('modal_root')||document.body,{childList:true,subtree:true});
})();
