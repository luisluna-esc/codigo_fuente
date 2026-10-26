/* Shared branding, toolbar and compact home layout. */
'use strict';
(()=>{
  const isHome=()=>currentPage==='dashboard'||currentPage==='music_module'&&SGL550.getSection()==='home';
  const cake='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 13h16v8H4zM3 17c2 2 4-2 6 0s4-2 6 0 4-2 6 0M7 9v4m5-4v4m5-4v4M7 4v2m5-3v3m5-2v2"/></svg>';
  function mission(){
    const existing=[...document.querySelectorAll('.music625-mission')];
    let footer=existing.find(el=>el.parentElement===document.body);
    if(!footer){footer=document.createElement('footer');footer.className='music625-mission';footer.textContent='Misión Boliviana Occidental Sur';document.body.append(footer);}
    for(const el of existing)if(el!==footer)el.remove();
  }
  function logo(){
    const image=document.getElementById('music615_onevoice');
    if(!image)return;
    const path='assets/onevoice-'+(document.documentElement.dataset.sgl545Mode==='dark'?'night':'day')+'.png';
    if(image.getAttribute('src')!==path)image.src=path;
  }
  function shell(){
    mission();
    const sidebar=document.querySelector('.sidebar'),bar=document.querySelector('.topbar');
    if(!sidebar||!bar||!currentUser())return;
    const title=sidebar.querySelector('.side-brand strong');
    if(title&&!title.querySelector('.music615-brand-line'))title.innerHTML='<span class="music615-brand-line">Sistema de Gestión</span><span class="music615-brand-line">de Música</span>';
    const version=sidebar.querySelector('.side-brand small');
    if(version)version.textContent='Versión '+String(APP_VERSION).split('-')[0];
    if(!sidebar.querySelector('.music615-brand-caption')){
      const caption=document.createElement('p');caption.className='music615-brand-caption';caption.textContent='Ministerio Joven MBOS';
      sidebar.querySelector('.side-brand')?.after(caption);
    }
    if(!bar.querySelector('#music615_onevoice')){
      const image=document.createElement('img');image.id='music615_onevoice';image.alt='ONEVOICE27 · MISIÓN PARA TODOS';image.width=294;image.height=28;
      bar.prepend(image);
    }
    logo();
    let birthday=bar.querySelector('#music615_birthday');
    if(SGL602.permit('birthdays.calendar')){
      if(!birthday){birthday=document.createElement('button');birthday.id='music615_birthday';birthday.type='button';birthday.className='btn ghost music615-top-icon';birthday.title='Ver cumpleaños';birthday.setAttribute('aria-label','Ver cumpleaños');birthday.innerHTML=cake;birthday.onclick=()=>window.sgl603ShowBirthdays?.();}
      const notice=bar.querySelector('#notification_pill');
      if(notice&&notice.nextElementSibling!==birthday)notice.after(birthday);
      else if(!birthday.isConnected)bar.querySelector('#sync_pill')?.before(birthday);
    }else birthday?.remove();
    window.SGL611Alerts?.refreshIcon();
  }
  function home(){
    const root=document.getElementById('content');if(!root)return;
    const home=isHome();root.classList.toggle('music615-home',home);
    if(!home)return;
    const hero=root.querySelector('.welcome-hero'),head=root.querySelector('.sgl551-music-head');
    if(head){
      head.classList.add('music615-home-heading');
      const title=head.querySelector('h1');if(title&&title.textContent.trim()!=='Inicio')title.textContent='Inicio';
      if(hero&&hero.nextElementSibling!==head)hero.after(head);
    }
    const leader=root.querySelector('.sgl551-leader-card'),ministry=root.querySelector('#music611_ministry_home');
    if(!leader||!ministry||root.querySelector('.music615-home-summary'))return;
    const quick=document.createElement('nav');quick.className='music615-home-shortcuts';quick.setAttribute('aria-label','Accesos de Inicio');
    const tasks=root.querySelector('.sgl562-shortcut[onclick*="tasks"]');
    if(tasks)quick.append(tasks);
    for(const section of ['events','certificates','investitures']){
      const tile=root.querySelector('.sgl551-option-card[onclick*="'+section+'"]');
      if(tile)quick.append(tile);
    }
    root.querySelector('.sgl562-shortcuts')?.remove();root.querySelector('.sgl551-option-grid')?.remove();
    if(head)head.after(quick);else hero?.after(quick);
    const summary=document.createElement('div');summary.className='music615-home-summary';
    leader.before(summary);summary.append(leader,ministry);
  }
  function decorate(){shell();home();}
  const content=setContent;setContent=function(...args){const result=content.apply(this,args);decorate();return result;};window.setContent=setContent;
  const previousShell=renderShell;renderShell=function(...args){const result=previousShell.apply(this,args);decorate();return result;};window.renderShell=renderShell;
  new MutationObserver(logo).observe(document.documentElement,{attributes:true,attributeFilter:['data-sgl545-mode']});
  window.MusicShell615={decorate,home,logo};
})();
