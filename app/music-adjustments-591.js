/* SGL 5.9.1 — Music-only discovery and progress scoring. */
(function(root){
 'use strict';
 const P='music_ministry';
 const normalize=s=>String(s||'').normalize('NFD').replace(/[\u0300-\u036f]/g,'').toLowerCase().trim();
 const matches=(name,query)=>normalize(query).split(/\s+/).every(part=>normalize(name).includes(part));
 function hasProgress(s){
  if(!s)return false;
  if(String(s.text||'').trim()||String(s.richText||'').replace(/<[^>]*>/g,'').trim()||s.primaryDocumentId||(s.attachmentIds||[]).length||(s.selections||[]).length||s.completionDate)return true;
  const values=x=>x&&typeof x==='object'?Object.values(x).some(values):typeof x==='string'?!!x.trim():x===true;
  return values(s.readingReport)||Object.keys(s.bibleReadings?.entries||{}).length>0;
 }
 function score(requirements,progress={}){
  const result={points:0,max:requirements.length*3,development:0,review:0,reviewed:0,total:requirements.length};
  for(const r of requirements){const status=progress.status?.[r.id]||'pending';if(status==='approved'){result.reviewed++;result.points+=3;}else if(status==='in_review'){result.review++;result.points+=2;}else if(['in_progress','needs_changes','assigned'].includes(status)&&hasProgress(progress.submissions?.[r.id])){result.development++;result.points++;}}
  result.percent=result.max?Math.round(result.points/result.max*100):0;result.reviewedPercent=result.total?Math.round(result.reviewed/result.total*100):0;return result;
 }
 root.SGL591Model={normalize,matches,hasProgress,score};
 if(typeof document==='undefined')return;
 const music=()=>!!root.__MUSICA_WEB__||root.SGL562?.view()==='mmus';
 const queries={aspirants:'',tutors:'',users:''};
 function search(){
  if(!music()||!['aspirants','tutors','users'].includes(currentPage))return;
  const host=document.getElementById('content');if(!host)return;
  const page=currentPage,getRows=()=>[...host.querySelectorAll(page==='tutors'?'.sgl562-tutors tbody tr,.sgl5060-tutor-row':'.table tbody tr,.sgl513r3-aspirant-card')].filter(r=>r.querySelector('strong,h3,td:not([colspan])'));
  if(host.querySelector('#music591_search')){host.querySelector('#music591_search').musicApply?.();return;}
  if(!getRows().length)return;
  const bar=document.createElement('section');bar.className='music592-search';bar.style.marginBottom='10px';
  bar.innerHTML='<label class="field-group" for="music591_search"><span>Buscar '+(page==='tutors'?'tutor':page==='users'?'usuario':'aspirante')+' por nombre</span><input id="music591_search" type="search" class="field" placeholder="Escribe parte del nombre…" autocomplete="off"></label><p class="muted" id="music591_results" role="status" aria-live="polite"></p>';
  const first=host.querySelector('.sgl562-tutors,.sgl5060-tutor-list,.table-wrap,.sgl513r3-aspirant-grid');if(first)first.before(bar);else host.prepend(bar);
  const input=bar.querySelector('input');input.value=queries[page];
  function apply(){queries[page]=input.value;let count=0;const rows=getRows();for(const row of rows){const label=row.querySelector('strong,h3')||row.querySelector('td:first-child'),name=label?.textContent||'',visible=matches(name,input.value);if(label&&input.value){const plain=normalize(name),hits=new Set();for(const part of normalize(input.value).split(/\s+/).filter(Boolean)){let pos=plain.indexOf(part);while(pos>=0){for(let i=pos;i<pos+part.length;i++)hits.add(i);pos=plain.indexOf(part,pos+part.length);}}label.innerHTML=[...name].map((ch,i)=>hits.has(i)?'<mark>'+esc(ch)+'</mark>':esc(ch)).join('');}else if(label&&label.querySelector('mark'))label.textContent=name;row.hidden=!visible;row.style.display=visible?'':'none';if(visible)count++;}bar.querySelector('#music591_results').textContent=count?`${count} de ${rows.length} coincidencias`:'No hay coincidencias.';}
  input.musicApply=apply;input.addEventListener('input',apply);apply();
 }
 const baseContent=setContent;setContent=function(...args){const result=baseContent.apply(this,args);search();return result;};root.setContent=setContent;
 const baseJourney=renderDailyActivity;
 renderDailyActivity=function(...args){
  if(!music())return baseJourney.apply(this,args);
  const actor=currentUser();if(!actor)return renderLogin();
  const personal=['aspirant','pastor_aspirant'].includes(actor.role)||root.SGL555?.personalMode();
  const people=(personal?[userById(actor.id)||actor]:accessibleAspirants(actor)).filter(u=>sgl421AccountActive(u)&&!u.musicAccessRemoved&&SGL555.personalPrograms(u).includes(P));
  const rows=people.map(u=>({u,s:score(v3ProgramRequirements(P),v3EnsureProgramProgress(u.id,P))})).sort((a,b)=>b.s.points-a.s.points||String(a.u.fullName).localeCompare(String(b.u.fullName),'es'));
  setTop(personal?'Mi ruta de Música':'Trayectoria de Música','El avance se calcula únicamente con informes REVISADOS');
  setContent(`<div class="page-head"><div><h1 class="page-title">${personal?'Mi ruta de Música':'Trayectoria de Música'}</h1><p class="subtitle">Solo los informes revisados hacen avanzar tu trayectoria.</p></div></div><div class="sgl-journey-grid">${rows.sort((a,b)=>b.s.reviewedPercent-a.s.reviewedPercent).map(({u,s})=>{const stat=sglJourneyStats(u.id,P);return `<article class="card sgl-journey-card"><span class="sgl-journey-badge ${esc(stat.level.id)}">${esc(stat.level.name||stat.level.label||stat.level.id)}</span><h2>${esc(u.fullName)}</h2><strong>${s.reviewedPercent}%</strong><progress max="100" value="${s.reviewedPercent}"></progress><p>${s.reviewed} de ${s.total} informes REVISADOS</p><p>${s.review} enviados a revisión · ${s.development} en desarrollo</p><div class="notice"><b>${stat.next?`${stat.need} informe${stat.need===1?'':'s'} revisado${stat.need===1?'':'s'} para ${esc(stat.next.label)}`:s.reviewed===s.total?'Programa completado':'Nivel Oro alcanzado'}</b></div><button class="btn primary" onclick="sglJourneyOpen('music_ministry',${esc(JSON.stringify(u.id))})">Ver requisitos</button></article>`;}).join('')||'<p>No hay aspirantes disponibles en tu alcance.</p>'}</div><details class="card"><summary>Puntos de Música</summary><p>1 punto por desarrollo con contenido, 2 por envío a revisión y 3 por informe revisado. Cada requisito cuenta según su estado, sin sumar puntos por navegar o abrir pantallas.</p><div class="table-wrap"><table class="table"><thead><tr><th>Nombre</th><th>Desarrollo</th><th>En revisión</th><th>Revisados</th><th>Puntos acumulados</th></tr></thead><tbody>${rows.map(({u,s})=>`<tr><td>${esc(u.fullName)}</td><td>${s.development}</td><td>${s.review}</td><td>${s.reviewed}</td><td>${s.points}</td></tr>`).join('')}</tbody></table></div></details>`);
 };root.renderDailyActivity=renderDailyActivity;
 const baseTutorSearch=renderTutors;renderTutors=function(...args){const r=baseTutorSearch.apply(this,args);search();return r;};root.renderTutors=renderTutors;
 root.SGL591={search,music};
})(typeof window==='undefined'?globalThis:window);

// Freeze the revision that was displayed; background polling must not approve unseen work.
(()=>{
 if(typeof openRequirement!=='function')return;
 const previous=openRequirement;
 openRequirement=async function(...args){
  window.sgl595DisplayedReview=null;
  const result=await previous.apply(this,args);
  if(document.querySelector('#modal_root .modal')&&modalOwnerId&&modalRequirementId){
   const sub=v3SubmissionFor(modalOwnerId,modalRequirementId,v3ModalProgramId);
   window.sgl595DisplayedReview={ownerId:modalOwnerId,programId:v3ModalProgramId,requirementId:modalRequirementId,updatedAt:String(sub?.updatedAt||''),submittedAt:String(sub?.submittedAt||'')};
  }
  return result;
 };window.openRequirement=openRequirement;
})();
