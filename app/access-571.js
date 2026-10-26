'use strict';
(()=>{
 const MUSIC='music_ministry', actual=()=>window.SGL555?.actual?.()||currentUser(),scope=()=>window.__MUSICA_WEB__||window.SGL562?.view?.()==='mmus'||v3ActiveProgramId===MUSIC?'music':'conquis';
 let drainedKey='',busy=false,draining=false,lastScope='',lastActor='',state=null,rows=[],inactive=new Set();
 const managers=['admin','departamental','secretaria','coordinador_musica'];
 const manager=()=>!SGL555?.personalMode()&&(actual()?.role==='admin'||scope()==='music'&&managers.includes(actual()?.role));
 window.sgl571CanResetMusic=(a,o)=>!!a&&!!o&&SGLMusicPolicy560.active(a)&&SGLMusicPolicy560.active(o)&&(v3ModuleRecord(MUSIC)?.assignedUserIds||[]).includes(o.id)&&(managers.includes(a.role)||['regional','tutor'].includes(a.role)&&a.id!==o.id&&sgl545HasTutor(o,a.id,MUSIC)||['aspirant','pastor_aspirant'].includes(a.role)&&a.id===o.id);
 window.sgl571DeleteMusicProgress=async function(ids,scope='requirement'){
  const a=currentUser(),o=userById(modalOwnerId)||targetAspirant();if(!sgl571CanResetMusic(a,o))return toast('No tienes permiso para eliminar este avance.');
  if(!confirm('¿Eliminar definitivamente el avance seleccionado? Los requisitos seguirán disponibles.'))return;
  const p=v3EnsureProgramProgress(o.id,MUSIC),body={actorId:a.id,ownerId:o.id,programId:MUSIC,scope,requirementId:scope==='requirement'?ids[0]:'',requirementIds:scope==='selection'?ids:[],operationId:newId('music_reset'),baseUpdatedAt:p.updatedAt||'',restoreGeneration:database.restoreGeneration||1};
  try{const {payload}=await v3Api('/api/progress/reset',{method:'POST',body,timeout:60000});if(!payload.ok)throw new Error(payload.error||'No se confirmó la eliminación.');await v3ApplyServerSnapshot(payload);await closeModal();render();toast('Avance eliminado en el servidor. El requisito quedó pendiente.')}catch(e){toast(e.status===409?'El avance cambió. Actualiza y vuelve a confirmar.':e.message||'No se confirmó la eliminación; el avance se conserva.')}
 };
 async function request(action='',extra={}){const opts={timeout:20000,body:action?{actorId:actual().id,action,...extra}:undefined,method:action?'POST':'GET'};const {payload}=await v3Api('/api/access571?actorId='+encodeURIComponent(actual().id)+'&programId='+(scope()==='music'?MUSIC:'class_grouped')+(extra.manage?'&manage=1':''),opts);return payload;}
 function shield(status,message=''){
  let root=document.getElementById('sgl571_access');if(!root){root=document.createElement('div');root.id='sgl571_access';root.style.cssText='position:fixed;inset:0;z-index:999999;background:#edf2f7;display:grid;place-items:center;padding:24px';document.body.append(root)}
  root.innerHTML=`<section class="card" style="max-width:580px;width:100%;padding:30px"><h1>${status.disabled?'ASPIRANTE INHABILITADO':status.scope==='conquis'?'ACCESO TEMPORALMENTE RESTRINGIDO':'BLOQUEO TEMPORAL'}</h1><p style="font-size:14px">${(status.reasons||[]).map(esc).join('<br>')}</p>${status.scope==='conquis'&&status.canExtend?'<p>Queremos animarte a continuar tu formación. Puedes habilitar tu acceso durante <strong>3 días</strong> para completar informes pendientes, realizar tareas y avanzar nuevos requisitos.</p>':''}${message?`<p class="notice">${esc(message)}</p>`:''}<div class="actions">${status.canExtend?'<button class="btn primary" id="sgl571_extend">Habilitar mi acceso por 3 días</button>':'<p>Comunícate con el administrador para habilitar tu acceso.</p>'}<button class="btn ghost" id="sgl571_check">Comprobar habilitación</button><button class="btn ghost" id="sgl571_leave">Salir</button></div></section>`;
  root.querySelector('#sgl571_extend')?.addEventListener('click',async()=>{try{await request('extend');root.remove();await check();if(!state?.blocked)renderShell()}catch(e){toast(e.message)}});
  root.querySelector('#sgl571_check').onclick=()=>check(true);
  root.querySelector('#sgl571_leave').onclick=async()=>{if(draining)return toast('Se está guardando el avance.');await logout();root.remove()};
 }
 async function finishWork(status){
  if(draining)return;draining=true;
  try{
   window.saveEmergencyDraft?.();await window.sgl5047SaveDraftOnly?.();
   const docSave=document.querySelector('#sgl5078_document_footer button[onclick*="sgl5078SaveDocument"]');if(docSave&&!docSave.disabled&&await window.sgl5078SaveDocument()!==true)throw new Error('El documento no terminó de guardarse. Se conserva abierto para reintentar.');
   if(document.getElementById('sgl531_form')&&await SGL531.save(false)!==true)throw new Error('La lectura no terminó de guardarse. Se conserva el borrador.');
   if(modalRequirementId&&document.querySelector('#req_text[contenteditable="true"]'))await saveRequirement(false);
   if(await closeModal()===false)throw new Error('Revisa el formulario abierto y guarda los cambios antes de cerrar.');
   await persist(false);shield(status,'Guardando los cambios pendientes en el servidor…');
   await syncNow();
   if(database.sync?.needsUpload||['pending','error','syncing'].includes(database.sync?.state))throw new Error('El borrador está guardado en este equipo. Falta confirmar el envío al servidor; vuelve a intentar con conexión.');
   await request('drained');drainedKey=lastActor+'|'+lastScope+'|'+status.blockedAt;shield(status,'Los cambios pendientes se guardaron. El acceso está temporalmente cerrado.');
  }catch(e){shield(status,e.message||'No se confirmó el guardado. Se conserva la copia local.')}finally{draining=false}
 }
 async function check(force=false){
  if(busy||!actual())return;busy=true;const who=actual().id,where=scope();if(who!==lastActor||where!==lastScope)inactive=new Set();
  try{const p=await request();if(actual()?.id!==who||scope()!==where)return;inactive=new Set(p.disabledIds||[]);state=p.status;lastActor=who;lastScope=where;
   if(state.blocked){if(drainedKey!==who+'|'+where+'|'+state.blockedAt||force)await finishWork(state)}else {document.getElementById('sgl571_access')?.remove();showFollowup();}
  }catch(e){if(e.status===423&&e.payload?.access571)shield(e.payload.access571)}finally{busy=false}
 }
 function showFollowup(){
  const host=document.getElementById('content');if(!host)return;
  if(scope()!=='conquis'||lastScope!=='conquis'||lastActor!==actual()?.id||state?.blocked){document.getElementById('sgl572_followup')?.remove();return}
  const pending=(state?.followup||[]).filter(x=>x.stage>0),until=state?.extensionUntil&&Date.parse(state.extensionUntil)>Date.now()?state.extensionUntil:'';
  if(!pending.length&&!until){document.getElementById('sgl572_followup')?.remove();return}
  const strong=pending.some(x=>x.stage===2);const text=until?'Tienes acceso habilitado hasta el '+new Date(until).toLocaleString()+'. Aprovecha este tiempo para desarrollar requisitos y presentar tus avances. Si necesitas orientación, comunícate con tu tutor.':strong?'En las últimas dos semanas registras avances o entregas pendientes. Te animamos a retomar tus informes, completar tus tareas y desarrollar nuevos requisitos. Cada avance cuenta y tu tutor puede acompañarte.':'Esta semana tienes entregas pendientes; aún puedes retomarlas. Tu tutor puede ayudarte a continuar tu formación.';
  const html='<strong>'+(until?'¡Es tu oportunidad de seguir avanzando!':'¡Sigamos avanzando en tu formación!')+'</strong><p>'+esc(text)+'</p><button class="btn primary small" onclick="SGL571.training()">'+(until?'Continuar mi formación':'Ver mis requisitos')+'</button>';
  let box=document.getElementById('sgl572_followup');if(!box){box=document.createElement('section');box.id='sgl572_followup';box.className='card notice';host.prepend(box)}if(box.innerHTML!==html)box.innerHTML=html;
 }
 async function training(){const pid=(state?.followup||[]).find(x=>x.stage>0)?.programId||(state?.followup||[])[0]?.programId;if(!pid)return;if(!['aspirant','pastor_aspirant'].includes(actual()?.role)){if(await SGL555.switchView('personal')===false)return}v3ActiveProgramId=pid;activeAspirantId=actual().id;go('requirements')}
 const oldContent=setContent;setContent=function(...args){const result=oldContent.apply(this,args);showFollowup();return result};window.setContent=setContent;
 async function manage(){
  if(!manager())return toast('Sin permiso para gestionar accesos.');
  const p=await request('',{manage:true});rows=p.rows||[];inactive=new Set(rows.filter(r=>r.disabled).map(r=>r.id));
  currentPage='access571';setTop('Accesos · '+(scope()==='music'?'Música':'Conquistadores'),'Bloqueos e inhabilitados');
  setContent(`<div class="page-head"><div><h1 class="page-title">Control de acceso</h1><p class="subtitle">${scope()==='conquis'?'Avisos y seguimiento, sin bloqueos automáticos. Selecciona a quién bloquear o habilitar. Durante el plazo de 3 días podrá trabajar en informes, tareas y nuevos requisitos.':'Bloqueo temporal, habilitación y aspirantes inhabilitados de Música.'}</p></div></div><div class="actions"><select id="sgl571_filter" class="field">${scope()==='conquis'?'<option value="followup">Necesitan seguimiento</option>':''}<option value="blocked">Bloqueados</option><option value="active">Activos</option><option value="disabled">Inhabilitados</option><option value="all">Todos</option></select><button class="btn ghost" onclick="SGL571.selectAll()">Seleccionar visibles</button><button class="btn red" onclick="SGL571.act('block')">Bloquear seleccionados</button><button class="btn primary" onclick="SGL571.act('release')">Habilitar seleccionados</button><button class="btn ghost" onclick="SGL571.act('disable')">Inhabilitar aspirantes</button><button class="btn ghost" onclick="SGL571.act('enable')">Reactivar</button><button class="btn red" onclick="SGL571.act('block',true)">Bloquear todos</button></div><div class="table-wrap"><table class="table"><thead><tr><th></th><th>Usuario</th><th>Cargo</th><th>Estado</th><th>${scope()==='conquis'?'Motivo / Seguimiento':'Motivo'}</th></tr></thead><tbody id="sgl571_rows"></tbody></table></div>`);
  document.getElementById('sgl571_filter').onchange=renderRows;renderRows();
 }
 function renderRows(){const filter=document.getElementById('sgl571_filter').value;document.getElementById('sgl571_rows').innerHTML=rows.filter(r=>filter==='all'||filter==='followup'&&(r.followup||[]).some(x=>x.stage>0)||filter==='disabled'&&r.disabled||filter==='blocked'&&r.blocked&&!r.disabled||filter==='active'&&!r.disabled&&!r.blocked).map(r=>`<tr><td>${r.id!==actual().id&&r.role!=='admin'?`<input type="checkbox" name="access571" value="${esc(r.id)}">`:''}</td><td>${esc(r.fullName)}</td><td>${esc(ROLE_LABELS[r.role]||r.role)}</td><td>${r.disabled?'Inhabilitado':r.blocked?'Bloqueado':'Habilitado'}</td><td>${(r.reasons||[]).map(esc).join('<br>')}${(r.followup||[]).map(x=>`<p><b>${esc(x.name)} · ${Math.round(x.percent)} % presentado</b><br>Tutor: ${esc(x.tutorName||'Sin asignar')}<br>Último avance: ${esc(x.lastProgressAt?new Date(x.lastProgressAt).toLocaleDateString():'Sin registro')} · Última entrega: ${esc(x.lastDeliveryAt?new Date(x.lastDeliveryAt).toLocaleDateString():'Sin registro')}<br>Entregas: ${x.deliveriesWeek} esta semana / ${x.deliveriesPreviousWeek} la anterior · Tareas: ${x.tasksDone}/${x.tasksTotal}<br>${(x.reasons||[]).map(esc).join(' ')}</p>`).join('')}${r.extensionUntil?'<p>Plazo: '+esc(new Date(r.extensionUntil).toLocaleString())+'</p>':''}</td></tr>`).join('')||'<tr><td colspan="5">No hay usuarios en este apartado.</td></tr>';}
 async function act(action,all=false){const ids=[...document.querySelectorAll('[name=access571]:checked')].map(x=>x.value);if(!all&&!ids.length)return toast('Selecciona al menos un usuario.');if(!confirm(`${action==='block'?'Bloquear':action==='disable'?'Inhabilitar':'Habilitar'} ${all?'todos los usuarios del módulo':ids.length+' usuarios'}?`))return;const reason=action==='block'?prompt('Motivo del bloqueo temporal:','Bloqueo temporal dispuesto por la administración.'):'';if(reason===null)return;try{await request(action,{ids,all,reason});await manage()}catch(e){toast(e.message)}}
 const baseNav=navItems;navItems=function(...args){const n=baseNav.apply(this,args);if(manager()&&!n.some(x=>x[0]==='access571'))n.push(['access571','⊘','Bloqueos / Inhabilitados']);return n};window.navItems=navItems;
 const baseGo=go;go=function(page,...args){if(page==='access571')return (async()=>{if(await closeModal()===false)return;window.__sgl555NavigationEpoch=Number(window.__sgl555NavigationEpoch||0)+1;return manage()})();return baseGo.call(this,page,...args)};window.go=go;
 const baseActive=sgl421AccountActive;sgl421AccountActive=function(u){return baseActive(u)&&!inactive.has(u?.id)};window.sgl421AccountActive=sgl421AccountActive;
 const basePeople=accessibleAspirants;accessibleAspirants=function(...args){return basePeople.apply(this,args).filter(u=>!inactive.has(u.id))};window.accessibleAspirants=accessibleAspirants;
 const baseShell=renderShell;renderShell=function(...args){const v=baseShell.apply(this,args);check();return v};window.renderShell=renderShell;
 const api=v3Api;v3Api=async function(path,...args){try{return await api.call(this,path,...args)}catch(e){if(e.status===423&&e.payload?.access571&&!String(path).startsWith('/api/access571'))shield(e.payload.access571);throw e}};window.v3Api=v3Api;
 // Self formation is distinct from supervising other leaders.
 const baseOpen=openRequirement;openRequirement=function(...args){if((v3ModalProgramId||v3ActiveProgramId)===MUSIC&&targetAspirant()?.id===actual()?.id&&!sgl545HasTutor(actual(),actual()?.tutorByProgram?.[MUSIC]||actual()?.tutorId,MUSIC))return toast('Necesitas un tutor asignado para registrar tu avance.');return baseOpen.apply(this,args)};window.openRequirement=openRequirement;
 const baseSwitch=SGL555.switchView;SGL555.switchView=async function(next,...args){if(next==='personal'&&scope()==='music'&&state?.trainingNeedsTutor)return toast('Necesitas un tutor de Música para realizar tu avance personal.');return baseSwitch.call(this,next,...args)};
 window.SGL571={training,isDisabled:id=>inactive.has(id),check,manage,act,selectAll:()=>document.querySelectorAll('[name=access571]').forEach(x=>x.checked=true)};
 setInterval(()=>{if(document.visibilityState!=='hidden')check()},15000);window.addEventListener('online',()=>check(true));
})();
