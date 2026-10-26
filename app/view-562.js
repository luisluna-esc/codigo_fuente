'use strict';
(()=>{
 const policy=SGLViewPolicy562,P=policy.MUSIC;
 const actor=()=>window.SGL555?.actual()||currentUser();
 const admin=()=>actor()?.role==='admin'&&!window.SGL555?.personalMode();
 let ownerKey='',selected='mc';
 function view(){if(window.__MUSICA_WEB__)return 'mmus';const id=actor()?.id||'';if(id!==ownerKey){ownerKey=id;selected='mc';try{selected=sessionStorage.getItem('sgl562_view_'+id)==='mmus'?'mmus':'mc';}catch{}}return selected;}
 const allowed=p=>!admin()||policy.program(p,view());
 const modules=()=>database.modules.filter(m=>m.active!==false&&allowed(m.id));
 const personal=u=>SGL555.personalPrograms(u);
 function inView(u){return policy.person(u,view(),personal(u),SGL555.managementPrograms(u),(database.classes||[]).some(c=>c.active!==false&&(c.assignedUserIds||[]).includes(u.id)));}
 function ensure(){if(!admin())return;if(!allowed(v3ActiveProgramId))v3ActiveProgramId=modules()[0]?.id||(view()==='mmus'?P:'class_grouped');document.documentElement.dataset.sglView=view();}
 async function select(next){
  if(!admin()||!['mc','mmus'].includes(next))return;
  if(window.__MUSICA_WEB__&&next==='mc'){location.href='/';return;}
  if(await closeModal()===false)return;
  selected=next;ownerKey=actor().id;try{sessionStorage.setItem('sgl562_view_'+ownerKey,next);}catch{}
  window.__sgl555NavigationEpoch=Number(window.__sgl555NavigationEpoch||0)+1;
  activeAspirantId=null;modalOwnerId=null;modalRequirementId=null;v3ModalProgramId=null;currentPage='dashboard';window.SGL551?.resetHistory();ensure();renderNav();renderDashboard();
 }
 const baseAssigned=v3AssignedProgramIds;v3AssignedProgramIds=function(u){const ids=baseAssigned.apply(this,arguments);return admin()?ids.filter(allowed):ids;};window.v3AssignedProgramIds=v3AssignedProgramIds;
 const basePeople=accessibleAspirants;accessibleAspirants=function(u=currentUser()){ensure();const people=basePeople.apply(this,arguments);return admin()?people.filter(p=>personal(p).some(allowed)):people;};window.accessibleAspirants=accessibleAspirants;
 const baseNav=navItems;navItems=function(u){ensure();const n=baseNav.apply(this,arguments);if(!admin())return n;if(view()!=='mmus')return policy.nav(n,view()).filter(([id])=>!['music_module','module_music_ministry'].includes(id)).map(([id,icon,label,...rest])=>[id,icon,id==='dashboard'?'Inicio':label,...rest]);const byId=new Map(n.map(item=>[item[0],item])),rows=[['dashboard','▦','Inicio'],['aspirants','◎','Aspirantes'],['reviews','✓','Informes/Revisión'],['music_events','▦','Eventos/Inscripciones'],['tasks','◷','Tareas'],['tutors','♙','Tutores'],['daily_activity','◇','Trayectoria'],['music_unassigned','♙','Aspirantes sin Tutor'],['users','☷','Usuarios'],['settings','⚙','Configuración'],['music_reports571','▤','Reportes'],['indications','✎','Configuración de Informes']];return rows.map(([id,icon,label])=>{const item=byId.get(id);return item?[id,icon,label,...item.slice(3)]:[id,icon,label];});};window.navItems=navItems;
 const baseGo=go;go=function(page,...args){ensure();if(admin()&&policy.nav([[String(page)]],view()).length===0)return toast('Esta opción no forma parte de Música.');return baseGo.call(this,page,...args);};window.go=go;
 const baseSelectProgram=v3SelectProgram;v3SelectProgram=function(p,...args){if(!allowed(p))return toast('Este módulo corresponde a la otra vista.');return baseSelectProgram.call(this,p,...args);};window.v3SelectProgram=v3SelectProgram;
 const baseMusicOpen=SGL550.open;SGL550.open=function(...args){if(admin()&&view()==='mc')return toast('Abre Inicio para acceder a Música.');return baseMusicOpen.apply(this,args);};
 const baseRender=render;render=function(){ensure();return baseRender.apply(this,arguments);};window.render=render;
 function homeIntro(){
  const musicContext=!!window.__MUSICA_WEB__||v3ActiveProgramId===P&&(!admin()||view()==='mmus');if(!admin()&&!musicContext)return '';const u=window.SGL555?.actual?.()||currentUser(),now=new Date(),greeting=now.getHours()<12?'Buenos días':now.getHours()<19?'Buenas tardes':'Buenas noches';
  const subtitle=musicContext?'Sistema de Gestión de Música · Formación y seguimiento.':'Conquistadores · Agrupada, Guía Mayor, Máster, Máster Avanzado y Clases.';
  const hero=typeof sgl5066WelcomeHeroMarkup==='function'?sgl5066WelcomeHeroMarkup(greeting,String(u?.fullName||'').trim().split(/\s+/)[0]||'',subtitle,now):'';
  const visible=new Set(navItems(currentUser()).map(n=>n[0]));
  const links=[['aspirants','◎','Aspirantes','Ver personas y expedientes'],['tutors','♙','Tutores','Consultar tutores y asignaciones'],['classes','◇','Clases','Planificación y seguimiento'],['clubs','◆','Clubes','Organizar clubes y sus logos'],['tasks','◷','Tareas','Revisar el trabajo pendiente'],['materials','▣','Materiales','Consultar recursos de apoyo'],['reviews','✓','Informes / Revisiones','Consultar entregas'],['users','☷','Usuarios','Gestionar cuentas de acceso']].filter(([id])=>visible.has(id));
  const shortcuts=links.map(([id,icon,title,desc])=>`<button type="button" class="card sgl562-shortcut" onclick="go('${id}')"><span class="sgl562-shortcut-icon" aria-hidden="true">${icon}</span><span><strong>${title}</strong><small>${desc}</small></span><span aria-hidden="true">›</span></button>`).join('');return hero+(shortcuts?'<section class="sgl562-shortcuts" aria-label="Accesos de Inicio">'+shortcuts+'</section>':'');
 }
 const baseDashboard=renderDashboard;renderDashboard=function(){ensure();if(!admin())return baseDashboard.apply(this,arguments);if(view()==='mmus')return SGL550.open('home');
  setTop('Inicio - Conquistadores','Administración');const people=accessibleAspirants(currentUser());
  setContent(`${homeIntro()}<div class="sgl562-section-heading"><h2>Módulos de Conquistadores</h2><span>${people.length} aspirantes</span></div><div class="grid3 program-dashboard-grid">${modules().map(m=>{const p=v3Program(m.id);return `<button type="button" class="card program-dashboard-card compact" onclick="SGL562.openProgram('${esc(m.id)}')"><div class="program-card-logo"><img src="${esc(p.logo||appLogo())}" alt=""></div><div><h3>${esc(p.menu_label||p.name)}</h3><p>${m.assignedUserIds?.length||0} aspirantes</p></div></button>`;}).join('')}<button type="button" class="card program-dashboard-card compact" onclick="go('classes')"><div><h3>Clases</h3><p>Miembros, planificación y asistencia</p></div></button></div>`);
 };window.renderDashboard=renderDashboard;
 function openProgram(p){if(!allowed(p))return;v3ActiveProgramId=p;openAspirantModulePreview(p);}
 const baseAspirants=renderAspirants;renderAspirants=function(){ensure();if(!admin()||view()!=='mc')return baseAspirants.apply(this,arguments);const people=accessibleAspirants(currentUser()).sort((a,b)=>String(a.fullName).localeCompare(String(b.fullName),'es'));setTop('Aspirantes · Vista Conquistadores','Administración');
  setContent(`<div class="page-head"><div><h1 class="page-title">Aspirantes</h1><p class="subtitle">Todos los aspirantes de los módulos Conquistadores, sin depender del último módulo abierto.</p></div><div class="actions"><button class="btn primary" onclick="openUser()">Crear aspirante</button></div></div><div class="table-wrap"><table class="table sgl562-people"><thead><tr><th>Nombres y Apellidos Completos</th><th>Región</th><th>Módulos Conquistadores</th><th>Acciones</th></tr></thead><tbody>${people.map(u=>`<tr><td>${esc(u.fullName)}</td><td>${esc(u.region||'—')}</td><td>${personal(u).filter(allowed).map(p=>esc(v3Program(p).menu_label||v3Program(p).name)).join(' · ')}</td><td><div class="actions"><button class="btn ghost small" onclick="SGL562.openAspirant('${esc(u.id)}')">Abrir expediente</button><button class="btn ghost small" onclick="openUser('${esc(u.id)}')">Editar accesos</button></div></td></tr>`).join('')||'<tr><td colspan="4">Sin aspirantes de Conquistadores.</td></tr>'}</tbody></table></div>`);
 };window.renderAspirants=renderAspirants;
 function openPerson(id){const u=userById(id);if(!u)return;const ids=personal(u).filter(allowed);if(!ids.length)return;if(!ids.includes(v3ActiveProgramId))v3ActiveProgramId=ids[0];return openAspirant(id);}
 const baseUsers=renderUsers;renderUsers=function(){ensure();if(!admin()||view()!=='mc')return baseUsers.apply(this,arguments);setTop('Usuarios Conquistadores','Administración');const people=database.users.filter(inView);
  setContent(`<div class="page-head"><div><h1 class="page-title">Cuentas de acceso</h1><p class="subtitle">Cuentas con formación o responsabilidades en Conquistadores.</p></div><div class="actions"><button class="btn primary" onclick="openUser()">Crear usuario</button><button class="btn ghost" onclick="SGL561.admin2()">Crear Administrador 2</button><button class="btn ghost" onclick="SGL561.template()">Plantilla Excel</button><label class="btn ghost">Importar Excel / CSV<input type="file" accept=".xlsx,.csv" hidden onchange="SGL561.importFile(this.files[0]);this.value=''"></label></div></div><div class="table-wrap"><table class="table sgl562-people"><thead><tr><th>Nombres y Apellidos Completos</th><th>Función</th><th>Región</th><th>Correo</th><th>Acciones</th></tr></thead><tbody>${people.map(u=>`<tr><td>${esc(u.fullName)}</td><td>${esc(ROLE_LABELS[u.role]||u.role)}</td><td>${esc(u.region||'—')}</td><td>${esc(u.email||u.username||'')}</td><td><div class="actions"><button class="btn ghost small" onclick="openUser('${esc(u.id)}')">Editar</button>${u.id!==actor().id?`<button class="btn ghost small" onclick="issueTemporaryPassword('${esc(u.id)}')">Clave temporal</button><button class="btn red small" onclick="deleteUserPermanent('${esc(u.id)}')">Eliminar</button>`:''}</div></td></tr>`).join('')}</tbody></table></div>`);
 };window.renderUsers=renderUsers;
 function tableTutors(){const root=document.getElementById('content');const list=root?.querySelector('.sgl5060-tutor-list');if(!list)return;const cards=[...list.querySelectorAll('.sgl5060-tutor-row')];if(!cards.length)return;const wrap=document.createElement('div');wrap.className='table-wrap sgl562-tutors';const table=document.createElement('table');table.className='table sgl562-people';table.innerHTML='<thead><tr><th>Tutor · Nombres y Apellidos Completos</th><th>Función / Región</th><th>Acciones</th></tr></thead><tbody></tbody>';const tbody=table.querySelector('tbody');
  for(const card of cards){const link=card.querySelector('[onclick*="sgl5060ShowTutorAspirants"]'),code=link?.getAttribute('onclick')||'',id=code.match(/sgl5060ShowTutorAspirants\(['"]([^'"]+)['"]\)/)?.[1],u=id?userById(id):null;if(admin()&&u&&!inView(u))continue;const tr=document.createElement('tr'),name=document.createElement('td'),meta=document.createElement('td'),actions=document.createElement('td'),group=document.createElement('div');group.className='actions';name.textContent=card.querySelector('strong')?.textContent||u?.fullName||'';meta.textContent=u?`${ROLE_LABELS[u.role]||u.role} · Región ${u.region||'—'}`:card.querySelector('p')?.textContent||'—';for(const b of card.querySelectorAll('button'))group.append(b);actions.append(group);tr.append(name,meta,actions);tbody.append(tr);}
  if(!tbody.children.length)tbody.innerHTML='<tr><td colspan="3">No hay tutores en esta vista.</td></tr>';wrap.append(table);list.replaceWith(wrap);
 }
 const baseTutors=renderTutors;renderTutors=function(){ensure();
  if(admin()&&view()==='mc'){
   const people=database.users.filter(u=>u.role==='tutor'&&sgl421AccountActive(u)&&SGL555.managementPrograms(u).some(allowed)).sort((a,b)=>String(a.fullName).localeCompare(String(b.fullName),'es'));
   setTop('Tutores · Vista Conquistadores','Administración');setContent(`<div class="page-head"><div><h1 class="page-title">Tutores</h1><p class="subtitle">Tutores de todos los módulos Conquistadores.</p></div><div class="actions"><button class="btn primary" onclick="openTutor()">Crear tutor</button></div></div><div class="table-wrap sgl562-tutors"><table class="table sgl562-people"><thead><tr><th>Nombres y Apellidos Completos</th><th>Región</th><th>Módulos Conquistadores</th><th>Acciones</th></tr></thead><tbody>${people.map(u=>`<tr><td>${esc(u.fullName)}</td><td>${esc(u.region||'—')}</td><td>${SGL555.managementPrograms(u).filter(allowed).map(p=>esc(v3Program(p).menu_label||v3Program(p).name)).join(' · ')}</td><td><div class="actions"><button class="btn ghost small" onclick="sgl5060ShowTutorAspirants('${esc(u.id)}')">Ver aspirantes asignados</button><button class="btn ghost small" onclick="openUser('${esc(u.id)}')">Editar</button></div></td></tr>`).join('')||'<tr><td colspan="4">Sin tutores de Conquistadores.</td></tr>'}</tbody></table></div>`);return;
  }
  const result=baseTutors.apply(this,arguments);tableTutors();return result;
 };window.renderTutors=renderTutors;
 const baseTutorMembers=window.sgl5060ShowTutorAspirants;window.sgl5060ShowTutorAspirants=function(id){if(!admin()||view()==='mmus')return baseTutorMembers.apply(this,arguments);const u=userById(id);if(!u||!inView(u))return;const people=accessibleAspirants(currentUser()).filter(p=>personal(p).filter(allowed).some(program=>sgl545HasTutor(p,id,program)));showModal(`<div class="modal small"><div class="modal-head"><h3>${esc(u.fullName)} · Aspirantes Conquistadores</h3><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="table-wrap"><table class="table sgl562-people"><tbody>${people.map(p=>`<tr><td>${esc(p.fullName)}</td><td><button class="btn ghost small" onclick="closeModal();SGL562.openAspirant('${esc(p.id)}')">Abrir expediente</button></td></tr>`).join('')||'<tr><td>Sin aspirantes asignados en Conquistadores.</td></tr>'}</tbody></table></div></div></div>`);};
 const baseRenderNav=renderNav;
 renderNav=function(){
  ensure();
  const result=baseRenderNav.apply(this,arguments);
  if(admin()){
   const dashboardButton=document.getElementById('nav')?.querySelector('[data-page="dashboard"]');
   if(dashboardButton){
    dashboardButton.setAttribute('onclick',"go('dashboard')");dashboardButton.setAttribute('aria-label','Inicio');
    dashboardButton.classList.toggle('active',currentPage==='dashboard'||currentPage==='music_module'&&SGL550.getSection()==='home');
   }
   if(view()==='mmus'){
    const sectionPage={home:'dashboard',readings:'music_readings',events:'music_events',certificates:'music_certificates',investitures:'music_investitures'};
    const page=currentPage==='music_module'?(sectionPage[SGL550.getSection()]||'dashboard'):currentPage;
    document.querySelectorAll('#nav button[data-page]').forEach(el=>el.classList.toggle('active',el.dataset.page===page));
   }
  }
  return result;
 };
 window.renderNav=renderNav;
 const baseContent562=setContent;setContent=function(html){const result=baseContent562.apply(this,arguments);queueMicrotask(()=>window.sgl562Birthdays?.());return result;};window.setContent=setContent;
 const baseSettings=renderSettings;renderSettings=function(){ensure();const presentation=document.getElementById('sgl571_user_movements')?sgl571CaptureView():null;const result=baseSettings.apply(this,arguments);if(admin()){const root=document.getElementById('content');for(const button of root.querySelectorAll('button')){const call=button.getAttribute('onclick')||'';if(call.includes('SGL561.directory(')){button.hidden=call.includes(view()==='mc'?"'music'":"'mc'");if(call.includes("'all'")){button.textContent=view()==='mc'?'VER USUARIOS Conquistadores':'VER USUARIOS MMus';button.setAttribute('onclick',`SGL561.directory('${view()==='mc'?'mc':'music'}')`);}}if(view()==='mc'&&/Copiar.*Música/i.test(button.textContent))button.hidden=true;}}sgl571RestoreView(presentation);return result;};window.renderSettings=renderSettings;
 const baseOpenUser=openUser;openUser=function(){const result=baseOpenUser.apply(this,arguments);if(admin()){for(const input of document.querySelectorAll('[name=user_module],[name=sgl555_management_program],[data-sgl545-program]')){const p=input.dataset.sgl545Program||input.value;if(!allowed(p)){const label=input.closest('label');if(label)label.hidden=true;}}}return result;};window.openUser=openUser;
 const baseDirectory=SGL561.directory;SGL561.directory=async function(group){if(!admin())return baseDirectory?.apply(this,arguments);await baseDirectory(view()==='mc'?'mc':'music');for(const button of document.querySelectorAll('#modal_root button')){const call=button.getAttribute('onclick')||'';if(call.includes('SGL561.directory(')){if(call.includes(view()==='mc'?"'music'":"'mc'"))button.hidden=true;if(call.includes("'all'"))button.hidden=true;}}};
 const css=document.createElement('link');css.rel='stylesheet';css.href='view-562.css?v='+(window.__MUSICA_WEB__?'1.0.24':'5.9');document.head.append(css);
 window.SGL562={homeIntro,view,select,allowed,filterPrograms:rows=>admin()?rows.filter(p=>allowed(p.id)):rows,inView,openProgram,openAspirant:openPerson,tableTutors};
})();

/* Connectivity is independent of the hidden page title and queue status. */
function sgl563ConnectionStatus({remote,online,verifiedOffline,state,message,configured}){
 const text=String(message||'');
 if((!remote&&online===false)||verifiedOffline===true||/sin conexi[oó]n|offline|servidor no disponible/i.test(text))return ['offline','Sin conexión'];
 if(verifiedOffline===false||state==='ok'||/conectado|conexión recuperada/i.test(text))return ['connected','Conectado'];
 if(!configured&&!remote)return ['offline','Sin conexión'];
 return ['connecting','Intentando conectar'];
}
(()=>{
 let reachable,attempting=false,sequence=0,settled=0;
 const api=v3Api;
 v3Api=async function(path,options={}){
  // Testing a different configured URL must not change this session's indicator.
  if(options.serverUrl&&options.serverUrl!==database?.settings?.serverUrl)return api.apply(this,arguments);
  const request=++sequence;if(reachable!==true){attempting=true;renderTopSync();}
  try{const result=await api.apply(this,arguments);if(request>=settled){settled=request;reachable=true;}return result;}
  catch(error){if(request>=settled){if(error.status){settled=request;reachable=true;}else if(error.name==='AbortError'||/fetch|network|alcanzar el servidor|conexión en espera/i.test(error.message||'')){settled=request;reachable=false;}}throw error;}
  finally{attempting=false;renderTopSync();}
 };window.v3Api=v3Api;
 const base=renderTopSync;
 renderTopSync=function(){
  const result=base.apply(this,arguments),pill=document.getElementById('sync_pill'),bar=document.querySelector('.topbar');
  if(!pill||!database)return result;
  if(bar&&pill.parentElement!==bar)bar.appendChild(pill);
  let [state,label]=sgl563ConnectionStatus({remote:REMOTE_WEB,online:navigator.onLine,verifiedOffline:reachable===undefined?(REMOTE_WEB?window.__SGL_REMOTE_VERIFIED_OFFLINE__:undefined):!reachable,state:database.sync?.state,message:reachable===true?'':database.sync?.message,configured:!!database.settings?.serverUrl});
  if(attempting&&state!=='connected'&&(REMOTE_WEB||navigator.onLine!==false)){state='connecting';label='Intentando conectar';}
  if(pill.textContent!==label)pill.textContent=label;
  pill.dataset.connection=state;pill.hidden=false;
  pill.setAttribute('aria-label','Estado de conexión: '+label+'. Ver sincronización');
  pill.title=[label,database.sync?.message||''].filter(Boolean).join('\n');
  pill.onclick=()=>window.sgl548RecoveryPanel();
  return result;
 };window.renderTopSync=renderTopSync;
 const shell=renderShell;renderShell=function(){const result=shell.apply(this,arguments);renderTopSync();return result;};window.renderShell=renderShell;
 window.addEventListener('offline',()=>renderTopSync());window.addEventListener('online',()=>renderTopSync());
})();
