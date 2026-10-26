/* Personal navigation and a single progress/tasks workspace. */
'use strict';
(()=>{
  const P='music_ministry';
  const personal=()=>['aspirant','pastor_aspirant'].includes(currentUser()?.role);
  const actual=()=>window.SGL555?.actual()||currentUser();
  const baseNav=navItems;
  navItems=function(user){
    const items=baseNav.call(this,user);
    if(!['aspirant','pastor_aspirant'].includes(user?.role))return items;
    const rows=[
      ['dashboard','▦','Inicio'],
      ['requirements','☷','Mi Tarjeta Líder'],
      ['music_readings','▤','Devoción Personal'],
      ['music_events','▦','Eventos/Inscripciones'],
      ['approved_reports','✓','Informes Revisados'],
      ['tasks','◇','Mi Avance/tareas'],
      ['music_certificates','▤','Certificados'],
      ['music_investitures','♙','Mi investidura']
    ];
    if(window.SGL555?.personalMode()&&SGL555.managementPrograms(actual()).length){
      rows.push(['my_management','⇄','Volver a Gestión']);
    }
    rows.push(['music_mentoring602','♡','Acompañamiento'],['profile','○','Mi Perfil'],['settings','⚙','Configuración']);
    return rows.filter(([id])=>id==='dashboard'||id==='my_management'||id==='music_investitures'||SGL602.permit('menu.'+id));
  };
  window.navItems=navItems;

  function progress(){
    if(!personal()||currentPage!=='tasks')return;
    const root=document.getElementById('content');
    if(!root)return;
    setTop('Mi Avance/tareas','Mi Tarjeta Líder de Música');
    const heading=root.querySelector('.page-head h1');
    if(heading)heading.textContent='Mi Avance/tareas';
    root.querySelector('#music611_my_progress')?.remove();
    root.querySelector('#music611_tasks_heading')?.remove();
    const s=SGL591Model.score(v3ProgramRequirements(P),v3EnsureProgramProgress(actual().id,P));
    const section=document.createElement('section');
    section.id='music611_my_progress';
    section.className='card music611-progress';
    section.innerHTML=`<h2>Mi avance</h2><strong>${Number(s.reviewedPercent)}%</strong>
      <progress max="100" value="${Number(s.reviewedPercent)}" aria-label="Mi avance revisado"></progress>
      <p>${Number(s.reviewed)} de ${Number(s.total)} requisitos revisados · ${Number(s.review)} en revisión · ${Number(s.development)} en desarrollo</p>
      <button type="button" class="btn ghost" onclick="go('requirements')">Ver mi tarjeta</button>`;
    const tasks=document.createElement('h2');tasks.id='music611_tasks_heading';tasks.textContent='Mis tareas';
    const head=root.querySelector('.page-head');
    if(head)head.after(section,tasks);else root.prepend(section,tasks);
  }
  function home(){
    if(!personal())return;
    const root=document.getElementById('content');
    const isHome=currentPage==='dashboard'||currentPage==='music_module'&&SGL550.getSection()==='home';
    if(!root||!isHome||root.querySelector('#music611_ministry_home'))return;
    const card=document.createElement('section');card.id='music611_ministry_home';card.className='card';
    card.innerHTML='<h2>Ministerios de Música</h2><p>Busca un ministerio, consulta sus datos y su historial.</p><button type="button" class="btn primary" onclick="go(\'music_ministries\')">Ver ministerios</button>';
    root.append(card);
  }
  const content=setContent;
  setContent=function(...args){const out=content.apply(this,args);home();progress();return out;};
  window.setContent=setContent;
  const tasks=renderTasks;
  renderTasks=function(...args){const out=tasks.apply(this,args);progress();return out;};
  window.renderTasks=renderTasks;
  const journey=renderDailyActivity;
  renderDailyActivity=function(...args){return personal()?renderTasks():journey.apply(this,args);};
  window.renderDailyActivity=renderDailyActivity;
  const navigate=go;
  go=function(page,...args){
    if(personal()&&page==='music_history610')return toast('El historial de gestiones está disponible en la vista de gestión.');
    if(personal()&&page==='daily_activity')page='tasks';
    if(personal()&&page==='music_investitures')return SGL550.open('investitures');
    if(personal()&&page==='requirements')return SGL550.requirements(actual().id);
    return navigate.call(this,page,...args);
  };
  window.go=go;
  window.SGL611Navigation={progress,home};
})();
