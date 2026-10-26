/* SGL 5.47 — interface only. Existing sync, outbox and review transport stay intact. */
'use strict';
const SGL547_PROGRAMS=['music_ministry'];
function sgl547CatalogPrograms(user=currentUser()){
  if(!user||user.active===false||['inactive','archived'].includes(user.accountState))return [];
  if(user.role==='admin')return [...SGL547_PROGRAMS];
  if(['aspirant','pastor_aspirant'].includes(user.role))return [];
  return SGL547_PROGRAMS.filter(id=>Array.isArray(user.requirementViewPrograms)&&user.requirementViewPrograms.includes(id));
}
function sgl547PendingTasks(user=currentUser()){
  const seen=new Set();
  return (database.tasks||[]).filter(task=>{
    if(!sgl506TaskIsRealForActor(task,user))return false;
    const pid=task.programId||'class_grouped';
    if(user?.role==='tutor'&&!sgl545HasTutor(userById(task.assigneeId),user.id,pid))return false;
    const key=[task.assigneeId,pid,task.requirementId].join('|');if(seen.has(key))return false;seen.add(key);
    const status=v3EnsureProgramProgress(task.assigneeId,pid).status?.[task.requirementId]||'pending';
    if(status==='needs_changes')return true;
    return !['in_review','approved'].includes(status)&&!task.completedAt&&!task.done&&!['completed','done'].includes(task.status);
  });
}
function sgl547UnreadNotices(user=currentUser()){
  return (database.syncNotifications||[]).filter(item=>sgl529NoticeRelevant(item,user)&&sgl504IsRealPendingNotification(item));
}
// Count unread decisions once per report, only while its current state matches.
function sgl547ReportDecisions(user=currentUser()){
  if(!['aspirant','pastor_aspirant'].includes(user?.role))return [];
  const seen=new Set();
  return sgl547UnreadNotices(user).filter(n=>{
    if(!['approved','needs_changes'].includes(n.type)||!n.requirementId)return false;
    const pid=n.programId||'class_grouped',key=pid+'|'+n.requirementId;
    const status=v3EnsureProgramProgress(user.id,pid).status?.[n.requirementId];
    if(status!==n.type||seen.has(key))return false;
    seen.add(key);return true;
  });
}
function sgl547WeekBirthdays(users,now=new Date()){
  const start=new Date(now.getFullYear(),now.getMonth(),now.getDate());
  start.setDate(start.getDate()-((start.getDay()+6)%7));
  const result=[];
  for(let offset=0;offset<7;offset++){
    const date=new Date(start);date.setDate(start.getDate()+offset);
    for(const person of users||[]){
      if(!['aspirant','pastor_aspirant'].includes(person.role)||person.active===false||['inactive','archived'].includes(person.accountState))continue;
      const parts=String(person.birthDate||'').match(/^\d{4}-(\d{2})-(\d{2})$/);
      if(parts&&Number(parts[1])===date.getMonth()+1&&Number(parts[2])===date.getDate())result.push({person,date});
    }
  }
  return result;
}
function sgl547BirthdayMessage(now=new Date()){
  return [
    ['¡Feliz cumpleaños!','Dios te regala un nuevo día y un nuevo año de vida. Camina con alegría y gratitud.','Inspirado en Salmo 118:24'],
    ['¡Celebremos tu vida!','Que Dios te bendiga, te guarde y te llene de paz. Tu servicio y tu vida son valiosos.','Inspirado en Números 6:24–26'],
    ['¡Un nuevo año para crecer!','Avanza con valentía: Dios te acompaña. Que este nuevo año renueve tu fuerza y tus sueños.','Inspirado en Josué 1:9'],
    ['¡Que tengas un cumpleaños bendecido!','Descansa con esperanza: Dios renueva tus fuerzas. Que su amor te sostenga en cada nuevo desafío.','Inspirado en Isaías 40:31']
  ][Math.floor(now.getHours()/6)];
}
// Preserve local presentation across a background data refresh, never across navigation.
function sgl571CaptureView(){
 const root=document.getElementById('content');if(!root)return null;
 const key=(el,i)=>el.id||el.name||String(i),fields=[];
 for(const [i,el] of [...root.querySelectorAll('input,textarea,select')].entries()){
  if(el.type==='file'||el.type==='password'||el.readOnly||el.disabled)continue;
  const check=['checkbox','radio'].includes(el.type),isSelect=el.tagName==='SELECT';
  const dirty=check?el.checked!==el.defaultChecked:isSelect?[...el.options].some(o=>o.selected!==o.defaultSelected):el.value!==el.defaultValue;
  if(dirty)fields.push({key:key(el,i),value:el.value,checked:el.checked,check,selected:isSelect?[...el.selectedOptions].map(o=>o.value):null});
 }
 return {page:currentPage,actor:currentUser()?.id,epoch:Number(window.__sgl555NavigationEpoch||0),details:[...root.querySelectorAll('details')].map((el,i)=>({key:key(el,i),open:el.open})),fields};
}
function sgl571RestoreView(state){
 if(!state||state.page!==currentPage||state.actor!==currentUser()?.id||state.epoch!==Number(window.__sgl555NavigationEpoch||0))return;
 const root=document.getElementById('content');if(!root)return;const key=(el,i)=>el.id||el.name||String(i);
 const details=new Map(state.details.map(v=>[v.key,v.open]));for(const [i,el] of [...root.querySelectorAll('details')].entries())if(details.has(key(el,i)))el.open=details.get(key(el,i));
 const fields=new Map(state.fields.map(v=>[v.key,v]));for(const [i,el] of [...root.querySelectorAll('input,textarea,select')].entries()){
  const old=fields.get(key(el,i));if(!old||el.type==='file'||el.type==='password')continue;
  if(old.check)el.checked=old.checked;else if(old.selected)for(const option of el.options)option.selected=old.selected.includes(option.value);else el.value=old.value;
 }
}
function sgl547ViewKey(){
  const keys=['users','modules','classes','clubs','tasks','materials','progress','programProgress','bibleYear','indicationRules','requirementConfigs','configs','reportRequirementModes','bibleVersions','conflicts','settings','music550'];
  const omit=new Set(['updatedAt','lastSeen','lastSeenAt','lastLoginAt','lastActiveAt','lastSyncedAt','serverToken','serverUrl','remoteSession','serverVersion','revision','lastLocalChangeAt']);
  const data={page:currentPage,program:v3ActiveProgramId,owner:activeAspirantId,actor:currentUser()?.id};
  const directory=['users','tutors'].includes(currentPage),settingsPage=currentPage==='settings';
  const settingKeys=new Set(['users','modules','classes','clubs','settings','configs','bibleVersions','conflicts','indicationRules','requirementConfigs','reportRequirementModes']);
  for(const k of keys){if(settingsPage&&!settingKeys.has(k))continue;if(directory&&!['users','modules','classes','settings'].includes(k))continue;data[k]=database[k];}
  if(settingsPage){data.audit=database.audit;data.errors=database.errors;}
  data.settings={...(database.settings||{}),activityHiddenIds:database.settings?.activityHiddenIds||[]};
  return JSON.stringify(data,(key,value)=>{if(omit.has(key))return undefined;return value;});
}
(()=>{
  let paintedKey='',catalogProgram='';
  const baseRelevant=sgl529NoticeRelevant;
  sgl529NoticeRelevant=function(item,user=currentUser()){
    if(!baseRelevant(item,user))return false;
    if(item?.type==='task_assigned')return sgl547PendingTasks(user).some(t=>t.assigneeId===user.id&&(t.programId||'class_grouped')===(item.programId||'class_grouped')&&t.requirementId===item.requirementId);
    return true;
  };window.sgl529NoticeRelevant=sgl529NoticeRelevant;
  const baseVisible=sgl504VisibleNotifications;
  sgl504VisibleNotifications=function(items){return baseVisible(items).filter(item=>sgl529NoticeRelevant(item));};window.sgl504VisibleNotifications=sgl504VisibleNotifications;
  const baseContent=setContent;
  setContent=function(html){const result=baseContent.apply(this,arguments);paintedKey=sgl547ViewKey();queueMicrotask(decorate);return result;};window.setContent=setContent;
  window.sgl547ViewUnchanged=()=>!!paintedKey&&paintedKey===sgl547ViewKey();

  const baseUser=openUser;
  openUser=function(id=''){
    const result=baseUser.apply(this,arguments);if(currentUser()?.role!=='admin')return result;
    const host=document.querySelector('.user-editor .field-row');if(!host)return result;
    const person=id?userById(id):null,box=document.createElement('fieldset');box.id='sgl547_permissions';box.className='field-group full sgl547-permissions';
    box.innerHTML='<legend>Consulta de requisitos por clase</legend><p class="muted">Acceso de lectura al catálogo. Solo el administrador puede asignarlo. No permite editar ni revisar informes ajenos.</p><div class="module-checkbox-grid">'+SGL547_PROGRAMS.map(pid=>`<label class="choice"><input type="checkbox" name="sgl547_view_program" value="${pid}" ${person?.requirementViewPrograms?.includes(pid)?'checked':''}><span>${esc(v3Program(pid).menu_label||v3Program(pid).name)}</span></label>`).join('')+'</div><button type="button" class="btn ghost small" data-all>Seleccionar las cinco</button>';
    box.querySelector('[data-all]').onclick=()=>box.querySelectorAll('input').forEach(x=>x.checked=true);host.appendChild(box);
    const toggle=()=>{box.hidden=['admin','aspirant','pastor_aspirant'].includes(value('user_role'));};document.getElementById('user_role')?.addEventListener('change',toggle);toggle();return result;
  };window.openUser=openUser;
  const baseItems=navItems;
  navItems=function(user){return baseItems.apply(this,arguments).filter(item=>!['notices','requirements_catalog'].includes(item[0]));};window.navItems=navItems;
  const baseGo=go;
  go=function(page,...args){
    if(page==='notices'){closeMobileMenu?.();return v4ShowNotifications(false);}
    if(page==='requirements_catalog'){
      if(!sgl547CatalogPrograms().length)return toast('El administrador debe habilitar la consulta de requisitos.');
      currentPage=page;renderNav();closeMobileMenu?.();renderCatalog();return;
    }
    closeMobileMenu?.();return baseGo.call(this,page,...args);
  };window.go=go;
  const baseRender=render;
  render=function(...args){if(currentPage==='requirements_catalog')return renderCatalog();return baseRender.apply(this,args);};window.render=render;
  function renderCatalog(){
    const ids=sgl547CatalogPrograms();if(!ids.length){currentPage='dashboard';return render();}
    if(!ids.includes(catalogProgram))catalogProgram=ids[0];
    setText('top_title','Consulta de requisitos');setText('top_eyebrow','Catálogo · solo lectura');
    const reqs=v3ProgramRequirements(catalogProgram);
    setContent(`<div class="page-head"><div><h2>Consulta de requisitos</h2><p class="subtitle">Clases habilitadas por el administrador · solo lectura</p></div></div><section class="card"><label class="field-group">Clase de liderazgo<select id="sgl547_catalog_program" class="field">${ids.map(id=>`<option value="${id}" ${id===catalogProgram?'selected':''}>${esc(v3Program(id).menu_label||v3Program(id).name)}</option>`).join('')}</select></label><label class="field-group">Buscar requisito<input id="sgl547_catalog_search" class="field" type="search" placeholder="Código, área o texto"></label><p class="muted">${reqs.length} requisitos · Este catálogo no muestra expedientes personales.</p><div id="sgl547_catalog_list">${reqs.map(req=>`<article class="sgl547-catalog-row"><small>${esc(req.areaTitle||req.area||'')} · ${esc(req.id)}</small><h3>${esc(req.official_text||req.title||req.id)}</h3>${req.subitems?.length?'<ul>'+req.subitems.map(x=>'<li>'+esc(typeof x==='string'?x:x.text||x.official_text||x.label||x.title||x.id)+'</li>').join('')+'</ul>':''}${req.selection?'<p class="muted">Selección: '+esc(req.selection.min||0)+' a '+esc(req.selection.max||req.subitems?.length||'')+' opciones.</p>':''}</article>`).join('')}</div></section>`);
    document.getElementById('sgl547_catalog_program').onchange=e=>{catalogProgram=e.target.value;renderCatalog();};
    document.getElementById('sgl547_catalog_search').oninput=e=>{const q=normalize(e.target.value);for(const row of document.querySelectorAll('.sgl547-catalog-row'))row.hidden=!normalize(row.textContent).includes(q);};
  }
  function badge(button,count){if(!button)return;let b=button.querySelector('.count');if(!count){b?.remove();return;}if(!b){b=document.createElement('i');b.className='count';button.appendChild(b);}if(b.textContent!==String(count))b.textContent=String(count);}
  window.sgl547RefreshBadges=function(){
    const user=currentUser();if(!user)return;const notices=sgl547UnreadNotices(user),pendingCount=sgl547PendingTasks(user).length;document.body.classList.toggle('sgl547-no-pending',pendingCount===0);
    for(const button of document.querySelectorAll('[data-page="tasks"]'))badge(button,pendingCount);
    for(const button of document.querySelectorAll('[data-page="notices"]'))badge(button,notices.length);
    if(['aspirant','pastor_aspirant'].includes(user.role))for(const pid of SGL547_PROGRAMS)for(const b of document.querySelectorAll(`[data-page="module_${pid}"]`))badge(b,notices.filter(n=>['approved','needs_changes','drive_reopened'].includes(n.type)&&(n.programId||'class_grouped')===pid).length);
    if(['aspirant','pastor_aspirant'].includes(user.role))for(const b of document.querySelectorAll('[data-page="approved_reports"], [data-page="reports"]')){
      const count=sgl547ReportDecisions(user).length;badge(b,count);
      b.querySelector('.count')?.setAttribute('aria-label',count+' informes revisados o que requieren corrección sin leer');
    }
    const menu=document.querySelector('#sgl547_dock [data-action="menu"]');badge(menu,notices.length);if(menu){menu.title=notices.length+' notificaciones sin leer';menu.setAttribute('aria-label','Menú · '+notices.length+' notificaciones sin leer');}
    const sync=document.querySelector('#sgl547_dock [data-action="sync"]');badge(sync,(database.syncQueue||[]).filter(op=>['queued','retry','sending'].includes(op.status)).length);
  };
  const baseNav=renderNav;
  renderNav=function(){
    const root=document.getElementById('nav'),user=currentUser();
    if(!root||!user)return baseNav.apply(this,arguments);
    const top=root.scrollTop,sidebar=root.closest('.sidebar'),sidebarTop=sidebar?.scrollTop||0;
    baseNav.apply(this,arguments);root.scrollTop=top;if(sidebar)sidebar.scrollTop=sidebarTop;
    window.sgl547RefreshBadges();decorateIcons();
  };window.renderNav=renderNav;
  const baseNotice=v4UpdateNotificationPill;
  v4UpdateNotificationPill=function(){const result=baseNotice.apply(this,arguments);window.sgl547RefreshBadges();return result;};window.v4UpdateNotificationPill=v4UpdateNotificationPill;
  const baseReminder=showAssignedTaskReminder;
  showAssignedTaskReminder=function(){if(!sgl547PendingTasks().length){taskReminderShown=true;return;}return baseReminder.apply(this,arguments);};window.showAssignedTaskReminder=showAssignedTaskReminder;

  // The Nexo reference is visual only. No Nexo data, branding or services are used.
  const paths={home:'M3 10 12 3l9 7 M5 9v12h5v-7h4v7h5V9',tasks:'M9 5h11v15H4V5h2 M9 3h6v4H9z M7 12l2 2 4-4 M7 17h9',book:'M12 5C9 3 5 3 3 4v16c3-1 6-1 9 1 3-2 6-2 9-1V4c-2-1-6-1-9 1v16',person:'M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0 M4 21v-3a8 8 0 0 1 16 0v3',menu:'M4 6h16 M4 12h16 M4 18h16',bell:'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9 M10 21h4',moon:'M20 15A9 9 0 0 1 9 3a9 9 0 1 0 11 12',sun:'M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0 M12 1v2 M12 21v2 M1 12h2 M21 12h2 M4 4l2 2 M18 18l2 2 M20 4l-2 2 M6 18l-2 2',chat:'M21 11a9 8 0 0 1-9 8H6l-4 3V11a9 8 0 0 1 19 0',folder:'M3 6h7l2 2h9v13H3z',chart:'M4 20V4 M4 20h17 M8 16v-4 M13 16V8 M18 16V5',settings:'M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0 M12 2v3 M12 19v3 M2 12h3 M19 12h3 M5 5l2 2 M17 17l2 2 M19 5l-2 2 M7 17l-2 2'};
  const icon=name=>`<svg class="sgl547-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="${paths[name]||paths.book}"/></svg>`;
  function pageIcon(page){if(page==='dashboard')return 'home';if(page==='tasks')return 'tasks';if(page==='notices'||page==='reviews')return 'bell';if(page==='profile'||page==='users'||page==='aspirants'||page==='tutors')return 'person';if(page==='settings')return 'settings';if(page==='daily_activity'||page==='journey')return 'chart';if(page==='materials'||page==='drive')return 'folder';return 'book';}
  function decorateIcons(){if(!matchMedia('(max-width:760px)').matches)return;for(const b of document.querySelectorAll('#nav [data-page] > b')){const next=icon(pageIcon(b.parentElement.dataset.page));if(b.innerHTML!==next)b.innerHTML=next;}for(const [id,name] of [['notification_pill','bell'],['message_pill','chat']]){const el=document.querySelector('#'+id+' > span:first-child');if(el&&el.innerHTML!==icon(name))el.innerHTML=icon(name);}}
  function modeIcon(){const b=document.getElementById('sgl545_mode');if(b&&matchMedia('(max-width:760px)').matches){const name=document.documentElement.dataset.sgl545Mode==='dark'?'sun':'moon';if(b.innerHTML!==icon(name))b.innerHTML=icon(name);}}
  const baseMode=window.sgl545ToggleMode;window.sgl545ToggleMode=function(){baseMode();modeIcon();};
  function decorate(){
    if(!currentUser()||!document.querySelector('.shell'))return;
    let dock=document.getElementById('sgl547_dock');
    if(!dock){dock=document.createElement('nav');dock.id='sgl547_dock';dock.className='sgl547-dock';dock.setAttribute('aria-label','Navegación móvil');document.body.appendChild(dock);}
    const items=navItems(currentUser()),quick=[['tasks','tasks','Tareas'],['profile','person','Perfil']].filter(([p])=>items.some(x=>x[0]===p));
    const key=JSON.stringify(quick);if(dock.dataset.items!==key){dock.dataset.items=key;dock.innerHTML=quick.map(([p,i,l])=>`<button type="button" data-page="${p}">${icon(i)}<span>${l}</span></button>`).join('')+`<button type="button" data-action="sync" aria-label="Estado de sincronización"><b class="sgl548-sync-letter">♫</b></button><button type="button" data-action="menu">${icon('menu')}<span>Menú</span></button><button type="button" data-action="logout" aria-label="Cerrar sesión"><svg class="sgl547-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><path d="M9 4H4v16h5M8 12h13m-5-5 5 5-5 5"/></svg><span>Salir</span></button>`;dock.onclick=e=>{const b=e.target.closest('button');if(!b)return;if(b.dataset.action==='sync'){window.sgl548RecoveryPanel();}else if(b.dataset.action==='logout'){logout();}else if(b.dataset.action==='menu'){e.stopPropagation();toggleMobileMenu();}else{closeMobileMenu();go(b.dataset.page);}};}
    for(const b of dock.querySelectorAll('[data-page]')){b.classList.toggle('active',b.dataset.page===currentPage);b.setAttribute('aria-current',b.dataset.page===currentPage?'page':'false');}
    document.body.classList.toggle('sgl547-modal-open',!!document.querySelector('#modal_root .modal-bg'));
    decorateIcons();modeIcon();window.sgl547RefreshBadges();
  }
  let birthdayPeople=null,birthdayFetching=false,birthdayFetchedAt=0,birthdayContext='';
  function daysToBirthday(person,now=new Date()){
    const parts=String(person.birthDate||'').match(/^\d{4}-(\d{2})-(\d{2})$/);if(!parts)return -1;
    const today=Date.UTC(now.getFullYear(),now.getMonth(),now.getDate());
    let year=now.getFullYear(),next=Date.UTC(year,Number(parts[1])-1,Number(parts[2]));
    if(next<today)next=Date.UTC(++year,Number(parts[1])-1,Number(parts[2]));
    return Math.round((next-today)/86400000);
  }
  window.sgl548DaysToBirthday=daysToBirthday;
  function birthdayView(){return window.__MUSICA_WEB__||v3ActiveProgramId==='music_ministry'?'mmus':'mc';}
  function birthdayKey(){return [currentUser()?.id,currentUser()?.role,window.SGL555?.personalMode(),birthdayView(),birthdayNow603().toDateString(),birthdayPermit603('birthdays.all'),birthdayPermit603('birthdays.calendar'),birthdayPermit603('birthdays.auto'),birthdayPermit603('birthdays.reminders')].join('|');}
  async function fetchBirthdays(){
    if(birthdayFetching||Date.now()-birthdayFetchedAt<60000)return;birthdayFetching=true;birthdayFetchedAt=Date.now();const key=birthdayContext;
    try{const {payload}=await v3Api('/api/birthdays?'+new URLSearchParams({actorId:currentUser().id,view:birthdayView()==='mmus'?'music':'conquistadores',mode:birthdayPermit603('birthdays.calendar')?'calendar':'today'}),{timeout:12000});if(key===birthdayKey()&&Array.isArray(payload?.people)){birthdayPeople=payload.people;birthdays();}}catch{}finally{birthdayFetching=false;}
  }
  function birthdayNow603(){const parts=new Intl.DateTimeFormat('en-CA',{timeZone:'America/La_Paz',year:'numeric',month:'2-digit',day:'2-digit'}).formatToParts(new Date());const value=k=>Number(parts.find(x=>x.type===k).value);return new Date(value('year'),value('month')-1,value('day'));}
  function birthdayPermit603(key){return birthdayView()!=='mmus'||window.SGL602?.permit(key)!==false;}
  function birthdayPeople603(){return birthdayPeople||window.SGLViewPolicy562?.birthdays(database.users||[],currentUser(),database,birthdayView(),birthdayNow603())||[];}
  function showBirthdays603(){if(!birthdayPermit603('birthdays.calendar'))return;const now=birthdayNow603(),people=birthdayPeople603().filter(p=>daysToBirthday(p,now)>=0).slice().sort((a,b)=>daysToBirthday(a,now)-daysToBirthday(b,now));showModal(`<div class="modal small"><div class="modal-head"><h3>Cumpleaños</h3><button class="modal-close" onclick="closeModal()" aria-label="Cerrar">×</button></div><div class="modal-body"><ul class="sgl548-birthday-list">${people.map(p=>`<li><span class="avatar">${esc(initials(p.fullName||''))}</span><span><strong>${esc(p.fullName||'')}</strong><small>${esc(p.birthDate.slice(5).split('-').reverse().join('/'))}</small></span></li>`).join('')||'<li>No hay fechas registradas.</li>'}</ul></div></div>`);}
  window.sgl603ShowBirthdays=async function(){
    if(!currentUser()||!birthdayPermit603('birthdays.calendar')||await closeModal()===false)return;
    const context=birthdayKey();
    if(context!==birthdayContext){birthdayContext=context;birthdayPeople=null;birthdayFetchedAt=0;}
    await fetchBirthdays();
    if(context===birthdayKey())showBirthdays603();
  };
  function birthdays(){
    if(currentPage!=='dashboard'&&!(currentPage==='music_module'&&window.SGL550?.getSection()==='home'))return;
    const root=document.getElementById('content'),user=currentUser();if(!root||!user)return;
    let tools=document.getElementById('sgm603_birthday_tools');
    if(birthdayPermit603('birthdays.calendar')){if(!tools){tools=document.createElement('div');tools.id='sgm603_birthday_tools';tools.className='sgm603-birthday-tools';tools.innerHTML='<button type="button" class="btn ghost small" data-birthdays>Ver cumpleaños</button>';root.append(tools);tools.querySelector('button').onclick=showBirthdays603;}const actions=root.querySelector('.music615-home-heading .actions');if(actions&&tools.parentElement!==actions)actions.append(tools);}else tools?.remove();
    if(!birthdayPermit603('birthdays.auto')&&!birthdayPermit603('birthdays.calendar')&&!birthdayPermit603('birthdays.reminders')){document.getElementById('sgl547_birthdays')?.remove();return;}
    const context=birthdayKey();if(context!==birthdayContext){birthdayContext=context;birthdayPeople=null;birthdayFetchedAt=0;}fetchBirthdays();
    const now=birthdayNow603(),people=birthdayPeople603(),today=people.filter(p=>daysToBirthday(p,now)===0);
    let box=document.getElementById('sgl547_birthdays');
    if(!today.length||!birthdayPermit603('birthdays.auto'))box?.remove();
    else{
      if(!box){box=document.createElement('section');box.id='sgl547_birthdays';box.className='card sgl547-birthdays';if(tools?.isConnected)tools.before(box);else root.append(box);}
      const key=JSON.stringify([now.toDateString(),today]);if(box.dataset.key!==key){box.dataset.key=key;box.innerHTML='<h3>Cumpleaños de hoy</h3><ul>'+today.map(p=>`<li><span class="avatar">${esc(initials(p.fullName||''))}</span><span><strong>${esc(p.fullName||'Usuario')}</strong><small>¡Feliz cumpleaños!</small></span></li>`).join('')+'</ul>';}
    }
    if(!birthdayPermit603('birthdays.reminders'))return;
    for(const p of people.filter(p=>[7,4].includes(daysToBirthday(p,now)))){
      const key='sgl548_birthday_'+user.id+'_'+p.id+'_'+now.toDateString();if(sessionStorage.getItem(key))continue;sessionStorage.setItem(key,'1');toast(p.fullName+' cumple años en '+daysToBirthday(p,now)+' días.');
    }
  }
  const baseDashboard=renderDashboard;
  renderDashboard=function(){const result=baseDashboard.apply(this,arguments);birthdays();return result;};window.renderDashboard=renderDashboard;
  window.sgl562Birthdays=birthdays;
  setInterval(()=>{if(!document.hidden)birthdays();},60000);
  const baseShell=renderShell;
  renderShell=function(){const result=baseShell.apply(this,arguments);decorate();paintedKey=sgl547ViewKey();return result;};window.renderShell=renderShell;
  let scheduled=false;
  new MutationObserver(()=>{if(scheduled)return;scheduled=true;requestAnimationFrame(()=>{scheduled=false;decoratePages();const hasModal=!!document.querySelector('#modal_root .modal-bg');document.body.classList.toggle('sgl547-modal-open',hasModal);});}).observe(document.body,{childList:true,subtree:true});
  window.addEventListener('resize',decorate);window.matchMedia('(prefers-color-scheme: dark)').addEventListener?.('change',modeIcon);

  // Explicit page selection: buttons refer to the page object, never the caret.
  function editor(){const e=document.getElementById('req_text');return e?.classList.contains('sgl5039-word-editor')&&e.dataset.sgl5039Editable!=='false'?e:null;}
  function pages(e){return [...e.querySelectorAll(':scope > .sgl5039-word-page')];}
  function decoratePages(){
    const e=editor(),bar=document.querySelector('.sgl537-page-actions');if(!e||!bar)return;
    let controls=bar.querySelector('.sgl547-page-tools');if(!controls){controls=document.createElement('div');controls.className='sgl547-page-tools';bar.appendChild(controls);}
    const ps=pages(e);
    ps.forEach((page,index)=>{
      let b=page.querySelector(':scope > .sgl547-page-delete-local');
      if(!b){b=document.createElement('button');b.type='button';b.className='sgl547-page-delete-local';b.dataset.sglEditorUi='1';b.contentEditable='false';b.addEventListener('pointerdown',event=>event.preventDefault());b.onclick=()=>confirmPage(e,page);page.appendChild(b);}
      const label='Eliminar página '+(index+1);if(b.textContent!=='▱−')b.textContent='▱−';b.setAttribute('aria-label',label);
      b.disabled=ps.length<2;b.title=b.disabled?'El informe debe conservar al menos una página':label;
    });
    if(controls.__pages?.length===ps.length&&controls.__pages.every((p,i)=>p===ps[i]))return;controls.__pages=ps;
    controls.replaceChildren();
    if(ps.length>1)for(const page of ps){const b=document.createElement('button');b.type='button';b.className='btn ghost small sgl547-delete-page';b.textContent='▱−';b.title='Eliminar página '+page.dataset.sgl5039Page;b.setAttribute('aria-label',b.title);b.addEventListener('pointerdown',event=>event.preventDefault());b.onclick=()=>confirmPage(e,page);controls.appendChild(b);}
    const undo=document.createElement('button');undo.type='button';undo.className='btn ghost small';undo.dataset.sgl547Undo='1';undo.textContent='Deshacer eliminación';undo.hidden=!e.__sgl547Undo;undo.onclick=()=>undoPage(e);controls.appendChild(undo);
  }
  function confirmPage(e,page){
    if(!e.contains(page)||pages(e).length<2||e.__sgl547PageDialog)return;
    const index=pages(e).indexOf(page)+1,body=page.querySelector('.sgl5039-page-body'),preview=body.cloneNode(true);preview.querySelectorAll('[data-sgl-editor-ui]').forEach(n=>n.remove());const text=preview.textContent.trim(),photos=body.querySelectorAll('figure[data-sgl-photo-id]').length;
    e.__sgl547PageDialog=true;page.classList.add('sgl547-page-target');
    const layer=document.createElement('div');layer.className='sgl547-page-confirm modal-bg';layer.setAttribute('role','dialog');layer.setAttribute('aria-modal','true');layer.setAttribute('aria-labelledby','sgl547_delete_title');
    layer.innerHTML=`<div class="modal small"><div class="modal-head"><h3 id="sgl547_delete_title">¿Eliminar la página ${index} de ${pages(e).length}?</h3></div><div class="modal-body"><p>${text||photos?'Se quitará el contenido de esta página:':'Esta página está vacía.'}</p>${text?'<blockquote>'+esc(text.slice(0,180))+'</blockquote>':''}${photos?'<p>'+photos+' fotografía(s).</p>':''}<p>Las demás páginas conservarán su contenido y se renumerarán. Puedes deshacer antes de seguir editando.</p></div><div class="modal-foot"><button type="button" class="btn ghost" data-cancel>Cancelar</button><button type="button" class="btn red" data-delete>Eliminar página ${index}</button></div></div>`;
    document.body.appendChild(layer);const previous=document.activeElement;
    const close=()=>{e.__sgl547PageDialog=false;page.classList.remove('sgl547-page-target');layer.remove();if(previous?.isConnected)previous.focus({preventScroll:true});};
    layer.querySelector('[data-cancel]').onclick=close;
    layer.querySelector('[data-delete]').onclick=async()=>{layer.querySelector('[data-delete]').disabled=true;try{await removePage(e,page);close();await window.sgl5039ReflowNow(true);decoratePages();}catch(error){close();toast('No se pudo eliminar la página: '+error.message);}};
    layer.addEventListener('keydown',event=>{if(event.key==='Escape'){event.preventDefault();event.stopPropagation();close();}if(event.key==='Tab'){event.preventDefault();const target=document.activeElement===layer.querySelector('[data-cancel]')?layer.querySelector('[data-delete]'):layer.querySelector('[data-cancel]');target.focus();}});
    layer.querySelector('[data-cancel]').focus();
  }
  async function removePage(e,page){
    if(!e.isConnected||!e.contains(page)||pages(e).length<2||editor()!==e)return;
    const ps=pages(e),index=ps.indexOf(page)+1,sub=v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);
    const before={html:window.sgl5039CanonicalHtml(e),attachments:[...(sub.attachmentIds||[])],boxes:structuredClone(sub.reportPhotoBoxes||{})};
    const removed=new Set([...page.querySelectorAll('figure[data-sgl-photo-id]')].map(f=>f.dataset.sglPhotoId));
    page.querySelector('.sgl5039-page-body').replaceChildren();
    // Deleting the first page removes the leading manual break on the next one.
    if(index===1)ps[1].querySelector('.sgl5039-page-body > [data-sgl-page-break]')?.remove();
    for(const f of e.querySelectorAll('figure[data-sgl-photo-id]')){const p=Number(f.dataset.sglPage||f.closest('.sgl5039-word-page')?.dataset.sgl5039Page||1);if(p>index){f.dataset.sglPage=String(p-1);const box=sub.reportPhotoBoxes?.[f.dataset.sglPhotoId];if(box)box.page=p-1;}}
    sub.attachmentIds=(sub.attachmentIds||[]).filter(id=>!removed.has(id));for(const id of removed)if(sub.reportPhotoBoxes)delete sub.reportPhotoBoxes[id];
    markEdit(e);e.__sgl547Undo=before;
  }
  function markEdit(e){window.sgl5047MarkReportDirty?.();e.querySelector('.sgl5039-page-body')?.dispatchEvent(new Event('input',{bubbles:true}));saveEmergencyDraft?.();}
  async function undoPage(e){
    const saved=e.__sgl547Undo;if(!saved||editor()!==e)return;
    const sub=v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);sub.attachmentIds=saved.attachments;sub.reportPhotoBoxes=saved.boxes;
    const ps=pages(e);for(const p of ps)p.querySelector('.sgl5039-page-body').replaceChildren();ps[0].querySelector('.sgl5039-page-body').innerHTML=saved.html;
    delete e.__sgl547Undo;markEdit(e);await window.sgl5039ReflowNow(true);decoratePages();
  }
  document.addEventListener('input',event=>{const e=editor();if(e&&e.contains(event.target)&&e.__sgl547Undo){delete e.__sgl547Undo;const b=document.querySelector('[data-sgl547-undo]');if(b)b.hidden=true;}},true);
  window.sgl547Decorate=decorate;window.sgl547DecoratePages=decoratePages;
})();
