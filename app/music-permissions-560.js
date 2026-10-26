/* Music-only UI for the complete AJUSTE FUNCIONES.xlsx permission matrix. */
'use strict';
(()=>{
 const P='music_ministry',policy=SGLMusicPolicy560;
 const music=()=>!!window.__MUSICA_WEB__||v3ActiveProgramId===P;
 const actual=()=>window.SGL555?.actual()||currentUser();
 const assigned=()=>v3ModuleRecord(P)?.assignedUserIds||[];
 const managing=()=>!window.SGL555?.personalMode()&&!['aspirant','pastor_aspirant'].includes(currentUser()?.role);
 const cap=key=>managing()&&policy.cap(currentUser(),key);
 const ownRegion=u=>currentUser()?.role!=='regional'||policy.sameRegion(currentUser(),u);
 const members=()=>accessibleAspirants(currentUser()).filter(u=>assigned().includes(u.id));
 function tutors(){
  const byId=new Map((database.musicTutorDirectory560||[]).map(u=>[u.id,u]));
  for(const u of database.users||[])if(policy.active(u)&&(['admin','departamental','secretaria','coordinador_musica','administrativo','regional','tutor','administrativo_sgm','lider_musica'].includes(u.role)&&SGL555.managementPrograms(u).includes(P)||members().some(o=>o.id!==u.id&&policy.tutorAssigned(o,u.id))))byId.set(u.id,{...byId.get(u.id),...u});
  return [...byId.values()].filter(u=>policy.active(u)&&(actual()?.role!=='regional'||u.createdBy===actual().id)&&(actual()?.role!=='administrativo'||u.createdBy===actual().id||members().some(o=>policy.tutorAssigned(o,u.id)))).sort((a,b)=>String(a.fullName||'').localeCompare(String(b.fullName||''),'es'));
 }
 const quote=x=>esc(JSON.stringify(String(x||'')));
 const btn=(text,code,primary=false)=>`<button type="button" class="btn ${primary?'primary':'ghost'}" onclick="${esc(code)}">${esc(text)}</button>`;
 const back=()=>'';
 const regions=selected=>'<option value="">Seleccionar región</option>'+[...Array.from({length:12},(_,i)=>String(i+1)),'MBOS'].map(r=>`<option value="${r}" ${r===policy.region(selected)?'selected':''}>${r}</option>`).join('');
 const fields=[['fullName','Nombres y Apellidos Completos','text'],['email','Correo','email'],['phone','Teléfono','tel'],['birthDate','Fecha de nacimiento','date'],['address','Dirección','text'],['church','Iglesia','text'],['missionaryDistrict','Distrito','text'],['mission','Misión','text'],['union','Unión','text']];
 function profileFields(u,editable){return fields.map(([key,label,type])=>`<label class="field-group">${label}<input class="field" type="${type}" name="${key}" value="${esc(u[key]||'')}" ${editable?'':'readonly'} ${key==='fullName'?'required':''}></label>`).join('')+`<label class="field-group">Región<select class="field" name="region" required ${editable?'':'disabled'}>${regions(u.region)}</select></label>`;}

 const baseReview=v3CanReviewRequirement;
 v3CanReviewRequirement=function(actor,owner,status='pending',...args){
  const program=typeof args[0]==='string'?args[0]:(v3ModalProgramId||v3ActiveProgramId);
  if(program!==P)return baseReview.apply(this,arguments);
  return status==='in_review'&&policy.review(actor,owner,assigned());
 };window.v3CanReviewRequirement=v3CanReviewRequirement;

 const baseDecision=reviewRequirement;
 reviewRequirement=function(){if((v3ModalProgramId===P||music())&&!policy.cap(currentUser(),'review'))return toast('Tu función no permite revisar ni solicitar correcciones en Música.');return baseDecision.apply(this,arguments);};window.reviewRequirement=reviewRequirement;
 function hideDeniedReviews(){if((v3ModalProgramId===P||music())&&!policy.cap(currentUser(),'review'))for(const button of document.querySelectorAll('#modal_root button[onclick*="reviewRequirement"],#modal_root .sgl5088-request-correction,#modal_root .sgl5088-approve'))if(!button.hidden)button.hidden=true;}
 const baseModal=showModal;
 showModal=function(){const result=baseModal.apply(this,arguments);hideDeniedReviews();return result;};window.showModal=showModal;
 new MutationObserver(hideDeniedReviews).observe(document.getElementById('modal_root')||document.body,{childList:true,subtree:true});

 const baseNav=navItems;
 navItems=function(u){
  let items=baseNav.apply(this,arguments);if(!music())return items;
  if(!cap('tutors'))items=items.filter(x=>x[0]!=='tutors');
  else if(!items.some(x=>x[0]==='tutors'))items.push(['tutors','♙','Tutores']);
  if(cap('instructions')&&!items.some(x=>x[0]==='indications'))items.push(['indications','✎','Indicaciones']);
  if(!cap('instructions'))items=items.filter(x=>x[0]!=='indications');
  if(!cap('registry'))items=items.filter(x=>!['music_registration','music_ministries'].includes(x[0]));
  else if(window.__MUSICA_WEB__&&!items.some(x=>x[0]==='music_ministries'))items.push(['music_ministries','♫','Ministerios de Música']);
  if(['aspirant','pastor_aspirant'].includes(u?.role)&&!(SGL550.complete(u)&&database.music550?.workflow?.[u.id]?.state==='approved'))items=items.filter(x=>x[0]!=='music_investitures');
  return items;
 };window.navItems=navItems;

 let showEmptyTutors=false;
 function toggleEmptyTutors(value){showEmptyTutors=!!value;renderTutors();}
 const tutorCount=id=>members().filter(u=>policy.tutorAssigned(u,id)).length;
 const baseTutors=renderTutors;
 renderTutors=function(){
  if(!music())return baseTutors.apply(this,arguments);
  if(!cap('tutors'))return go('dashboard');
  const visibleTutors=tutors().filter(u=>showEmptyTutors||tutorCount(u.id)>0);
  setTop('Tutores de Música',ROLE_LABELS[currentUser().role]);
  setContent(`<div class="page-head"><div><h1 class="page-title">Tutores de Música</h1><p class="subtitle">${currentUser().role==='regional'?'Solo tutores creados por ti':'Personas que ejercen como tutores en Música, conservando su cargo.'}</p></div><div class="actions">${back()}${cap('createTutor')?'<button type="button" class="btn primary" id="sgl556_create_tutor" onclick="SGL556.openTutor()">+ Crear tutor</button>':''}</div></div><label class="sgm601-tutor-filter"><input type="checkbox" ${showEmptyTutors?'checked':''} onchange="SGL560.toggleEmptyTutors(this.checked)"> Mostrar también tutores sin aspirantes en mi alcance</label><div class="sgl5060-tutor-list">${visibleTutors.map(u=>`<article class="card sgl5060-tutor-row"><div><strong>${esc(u.fullName)}</strong><p>${esc(ROLE_LABELS[u.role]||u.role)} · Región ${esc(policy.region(u.region)||'sin registrar')}</p></div>${btn('Ver aspirantes ('+tutorCount(u.id)+')',`sgl5060ShowTutorAspirants(${JSON.stringify(u.id)})`)}${actions(u)}</article>`).join('')||'<article class="card">No hay tutores con este filtro en tu alcance.</article>'}</div>`);
 };window.renderTutors=renderTutors;
 const baseAssigned=window.sgl5060ShowTutorAspirants;
 window.sgl5060ShowTutorAspirants=function(id){
  if(!music())return baseAssigned.apply(this,arguments);
  const tutor=tutors().find(u=>u.id===id);if(!cap('tutors')||!tutor)return toast('Tutor fuera de tu alcance.');
  const people=members().filter(u=>policy.tutorAssigned(u,id));
  showModal(`<div class="modal small"><div class="modal-head"><h3>${esc(tutor.fullName)}</h3><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><p>Aspirantes asignados que puedes consultar.</p><div class="list">${people.map(u=>`<article class="card"><div class="actions">${btn(u.fullName,`SGL550.requirements(${JSON.stringify(u.id)});closeModal()`)}${managing()&&policy.remove(currentUser(),u)&&window.SGL602?.permit('people.remove')!==false?btn('Eliminar de Música',`SGL560.removeFromTutor(${JSON.stringify(u.id)})`):''}</div></article>`).join('')||'<p>No hay aspirantes asignados en tu alcance.</p>'}</div></div></div>`);
 };

 function actions(u){return `<div class="actions">${window.SGL598?.contactMarkup(u)||''}${assigned().includes(u.id)?btn('♫ Ver expediente',`SGL550.requirements(${JSON.stringify(u.id)})`):''}${cap('profile')?btn('✎ Editar datos',`SGL560.profile(${JSON.stringify(u.id)})`):''}${currentUser()?.role==='admin'?btn('Configurar cuenta',`openUser(${JSON.stringify(u.id)})`):''}${managing()&&policy.remove(currentUser(),u)?btn('Eliminar de Música',`SGL560.remove(${JSON.stringify(u.id)})`):''}${['admin','departamental','secretaria','coordinador_musica','administrativo_sgm','administrativo','regional'].includes(currentUser()?.role)&&assigned().includes(u.id)?btn('Asignar / cambiar tutor',`SGL561.assign(${JSON.stringify(u.id)})`):''}</div>`;}
 const baseUsers=renderUsers,baseAspirants=renderAspirants;
 function peopleView(users=false){
  const list=members().filter(u=>users||!!(Object.hasOwn(u,'tutorByProgram')?u.tutorByProgram?.[P]:u.tutorId));
  if(users&&cap('profile'))for(const u of database.users||[])if(!list.some(o=>o.id===u.id)&&policy.active(u)&&(assigned().includes(u.id)||window.SGL555?.managementPrograms(u).includes(P)))list.push(u);
  list.sort((a,b)=>String(a.fullName||'').localeCompare(String(b.fullName||''),'es')||String(a.id).localeCompare(String(b.id)));
  setTop(users?'Usuarios de Música':'Aspirantes de Música',ROLE_LABELS[currentUser().role]);
  setContent(`<div class="page-head"><div><h1 class="page-title">${users?'Usuarios':'Aspirantes'} de Música</h1><p class="subtitle">${currentUser().role==='regional'?'Aspirantes asignados a ti y los que creaste mientras estén sin tutor.':'Personas y expedientes dentro de tu alcance.'}</p></div><div class="actions">${back()}${managing()?btn('＋ Crear aspirante',"SGL560.createAspirant()",true):''}${currentUser().role==='admin'&&users?btn('Crear otra función',"openUser()") : ''}</div></div><div class="table-wrap"><table class="table"><thead><tr><th>Nombre</th><th>Función</th><th>Región</th><th>Acciones</th></tr></thead><tbody>${list.map(u=>`<tr data-user-id="${esc(u.id)}"><td><strong>${esc(u.fullName)}</strong></td><td>${esc(ROLE_LABELS[u.role]||u.role)}</td><td>${esc(policy.region(u.region)||'Sin registrar')}</td><td>${actions(u)}</td></tr>`).join('')||'<tr><td colspan="4">No hay personas en tu alcance.</td></tr>'}</tbody></table></div>`);
 }
 renderUsers=function(){if(!music())return baseUsers.apply(this,arguments);return peopleView(true);};window.renderUsers=renderUsers;
 renderAspirants=function(){if(!music())return baseAspirants.apply(this,arguments);return peopleView(false);};window.renderAspirants=renderAspirants;

 async function account(action,u,value={}){
  const actor=actual(),epoch=window.__sgl555NavigationEpoch;
  const {payload}=await v3Api('/api/accounts/music',{method:'POST',headers:{'X-SGL-Personal':'0'},timeout:20000,body:{actorId:actor.id,ownerId:u?.id||'',action,value,expectedUpdatedAt:u?.updatedAt||'',restoreGeneration:database.restoreGeneration||1,programId:P}});
  if(!payload.ok)throw new Error(payload.error||'No se confirmó el cambio.');
  const {payload:snap}=await v3Api('/api/accounts/view?actorId='+encodeURIComponent(actor.id),{timeout:15000});
  const applied=await v3ApplyServerSnapshot(snap);if(applied?.stale||actual()?.id!==actor.id)return false;
  if(epoch===window.__sgl555NavigationEpoch){renderNav();render();}
  return true;
 }
 async function remove(id){return SGL561.remove(id);}
 async function removeFromTutor(id){if(await closeModal()===false)return;return SGL560.remove(id);}
 async function assignSelf(){return toast('Solicita la asignación desde la lista Sin tutor.');}

 function profile(id){const u=userById(id);if(!u||(u.id!==actual()?.id&&!policy.cap(actual(),'profile')))return;showModal(`<div class="modal small"><div class="modal-head"><h3>Datos personales de Música</h3><button class="modal-close" onclick="closeModal()">×</button></div><form id="sgl560_profile_form"><div class="modal-body"><div class="field-row">${profileFields(u,true)}</div><p id="sgl560_profile_error" role="status"></p></div><div class="modal-foot"><button class="btn primary">Guardar datos</button><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button></div></form></div>`);document.getElementById('sgl560_profile_form').onsubmit=async e=>{e.preventDefault();const button=e.target.querySelector('button');button.disabled=true;try{const data=Object.fromEntries(new FormData(e.target));if(String(data.email||'').trim().toLowerCase()!==String(u.email||'').trim().toLowerCase()){if(!confirm('¿Estás seguro de modificar tu correo? Cambiará el correo asociado a tu cuenta y su inicio de sesión.'))return;data.emailChangeConfirmed=true;}await account('profile',u,data);await closeModal();toast('Datos guardados.');}catch(error){setText('sgl560_profile_error',error.message);}finally{button.disabled=false;}};}
 function createAspirant(){
  if(!managing())return;const actor=currentUser(),selected=actor.role==='regional'?actor.region:'';
  const options=[...new Map([...(database.musicTutorDirectory560||[]),...(database.users||[])].map(u=>[u.id,u])).values()].filter(u=>policy.active(u)&&(u.musicIsTutor560||window.SGL555?.managementPrograms(u).includes(P))&&(!['administrativo','regional'].includes(actor.role)||['administrativo','regional','tutor','lider_musica'].includes(u.role)));
  showModal(`<div class="modal small"><div class="modal-head"><h3>Crear aspirante de Música</h3><button class="modal-close" onclick="closeModal()">×</button></div><form id="sgl560_create_form"><div class="modal-body"><div class="field-row">${profileFields({region:selected},true)}<label class="field-group">Usuario<input class="field" name="username" required></label><label class="field-group">Contraseña temporal<input class="field" name="password" type="password" minlength="8" required></label><label class="field-group">Función<select class="field" name="role"><option value="aspirant">Líder aspirante</option><option value="pastor_aspirant">Pastor aspirante</option></select></label><label class="field-group">Tutor de Música<select class="field" name="tutorId" ${actor.role==='regional'?'':'required'}>${['admin','departamental','secretaria','coordinador_musica','administrativo_sgm'].includes(actor.role)?`<option value="${esc(actor.id)}">${esc(actor.fullName)} (yo)</option>`:''}<option value="">${actor.role==='regional'?'Sin tutor':'Selecciona un tutor'}</option>${options.map(u=>`<option value="${esc(u.id)}">${esc(u.fullName)}</option>`).join('')}</select></label></div><p id="sgl560_create_error" role="status"></p></div><div class="modal-foot"><button class="btn primary">Crear aspirante</button></div></form></div>`);
  const form=document.getElementById('sgl560_create_form');if(actor.role==='regional')form.elements.region.disabled=true;
  form.onsubmit=async e=>{e.preventDefault();const button=form.querySelector('button');button.disabled=true;try{const data=Object.fromEntries(new FormData(form));data.region=actor.role==='regional'?policy.region(actor.region):data.region;data.id=newId('usr');data.password=await passwordRecord(data.password);await account('create_aspirant',null,data);await closeModal();toast('Aspirante creado con su tarjeta de Música.');}catch(error){setText('sgl560_create_error',error.message);}finally{button.disabled=false;}};
 }

 const baseProfile=renderProfile;
 renderProfile=function(){
  if(!music())return baseProfile.apply(this,arguments);
  const u=actual(),editable=true,tutorId=Object.hasOwn(u,'tutorByProgram')?u.tutorByProgram?.[P]:u.tutorId,tutor=userById(tutorId);
  setTop('Mi perfil',ROLE_LABELS[u.role]);
  setContent(`<div class="page-head"><div><h1 class="page-title">Mi perfil</h1><p class="subtitle">${editable?'Datos personales de tu cuenta.':'Datos registrados de tu cuenta.'}</p></div>${back()}</div><article class="card profile"><div><div class="profile-photo">${avatarMarkup(u)}</div><label class="btn ghost">Cargar foto<input type="file" accept="image/png,image/jpeg" hidden onchange="SGL592.photo(this)"></label></div><div><div class="field-row">${profileFields(u,false)}</div>${tutor?`<p>Mi tutor de Música: <strong>${esc(tutor.fullName)}</strong></p>`:''}<div class="actions">${editable?btn('✎ Editar mis datos',`SGL560.profile(${JSON.stringify(u.id)})`,true):''}${btn('Cambiar contraseña',"changeOwnPassword()")}</div></div></article>`);
 };window.renderProfile=renderProfile;

 const baseOpenUser=openUser;
 openUser=function(id=''){if(music()&&id&&!(managing()&&currentUser()?.role==='admin'))return policy.cap(actual(),'profile')?profile(id):toast('Tu función no permite editar datos personales.');return baseOpenUser.apply(this,arguments);};window.openUser=openUser;
 const baseSaveUser=saveUser;
 saveUser=async function(e,id=''){if(music()){const region=document.getElementById('user_region');if(region&&!policy.region(region.value)){e?.preventDefault();return setText('user_error','Selecciona una región: 1–12 o MBOS.');}}return baseSaveUser.apply(this,arguments);};window.saveUser=saveUser;
 const baseDelete=window.deleteUser,basePermanent=window.deleteUserPermanent;
 window.deleteUser=function(id){if(music())return remove(id);return baseDelete?.apply(this,arguments);};
 window.deleteUserPermanent=function(id){if(music())return remove(id);return basePermanent?.apply(this,arguments);};

 // The generic instruction editor is narrowed to the current Music program.
 const baseContent=setContent;
 setContent=function(){const result=baseContent.apply(this,arguments);if(music()){const select=document.getElementById('sgl548_indication_program');if(select){for(const option of [...select.options])if(option.value!==P)option.remove();select.value=P;}if(currentUser()?.role==='regional'){const region=document.querySelector('#sgl556_tutor_form [name="region"]');if(region){region.value=policy.region(currentUser().region);region.disabled=true;}}}return result;};window.setContent=setContent;
 const baseRenderNav=renderNav;renderNav=function(){const result=baseRenderNav.apply(this,arguments);window.sgl548ApplyTheme?.();return result;};window.renderNav=renderNav;
 if(window.MusicaApp)window.MusicaApp.removeMember=remove;
 window.SGL560={account,tutors,tutorCount,toggleEmptyTutors,profile,remove,removeFromTutor,assignSelf,createAspirant,members,policy};
})();
