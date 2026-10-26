/* SGL 5.47 — diarios personales de Sistema de Gestión de Música.
 * El contenido viaja en submission.bibleReadings, campo admitido por R3.
 * Calendario: 25/09/2026 = Salmos 49, revivalandreformation.org/bhp/en/bible/psa/49.
 */
(function(root){
  'use strict';
  const PROGRAM='music_ministry',YEAR='LM-NUT-1',FORTY='LM-TEM-1',KIND='music_devotion',DAY=86400000;
  const clone=value=>JSON.parse(JSON.stringify(value));
  function dateValue(value){
    if(!/^\d{4}-\d{2}-\d{2}$/.test(String(value)))return NaN;
    const n=Date.parse(value+'T00:00:00Z');return Number.isFinite(n)&&new Date(n).toISOString().slice(0,10)===value?n:NaN;
  }
  const addDays=(date,n)=>new Date(dateValue(date)+n*DAY).toISOString().slice(0,10);
  function annualDates(start){
    if(!Number.isFinite(dateValue(start)))return [];
    const end=new Date(dateValue(start));end.setUTCFullYear(end.getUTCFullYear()+1);
    // Un aniversario del 29 de febrero se celebra el 28 de febrero.
    if(start.endsWith('-02-29'))end.setUTCDate(0);
    return Array.from({length:Math.round((end-dateValue(start))/DAY)},(_,i)=>addDays(start,i));
  }
  const canonCache=new WeakMap();
  function chapter(date,canon){
    if(!Number.isFinite(dateValue(date)))throw new Error('Fecha inválida.');
    let all=canonCache.get(canon);if(!all){all=canon.flatMap(([code,name,count])=>Array.from({length:count},(_,i)=>({code,name,chapter:i+1,reference:`${name} ${i+1}`})));canonCache.set(canon,all);}
    const anchor=all.findIndex(x=>x.code==='PSA'&&x.chapter===49);
    if(anchor<0)throw new Error('Falta el canon bíblico.');
    const index=anchor+Math.round((dateValue(date)-dateValue('2026-09-25'))/DAY);
    return all[((index%all.length)+all.length)%all.length];
  }
  function entryErrors(kind,entry,today){
    const errors=[];
    if(!Number.isFinite(dateValue(entry.date))||entry.date>today)errors.push('Indica la fecha real de lectura, hasta hoy.');
    const fields=[['favoriteVerses','Versículo favorito'],['impactStory','Lo que más me impactó'],['lifeLesson','Aplicación para mi vida']];
    if(kind==='forty'&&!entry.songTheme&&entry.song&&entry.songRelation)entry={...entry,songTheme:entry.song+' — '+entry.songRelation};
    if(kind==='forty')fields.push(['theme','Tema de adoración'],['reference','Pasaje bíblico'],['songTheme','Himno o canto y cómo se relaciona con el tema']);
    for(const [key,label] of fields)if(!String(entry[key]||'').trim())errors.push(`Completa: ${label}.`);
    if(entry.time&&!/^([01]\d|2[0-3]):[0-5]\d$/.test(entry.time))errors.push('La hora no es válida.');
    return errors;
  }
  function keysFor(req,data){return req===FORTY?Array.from({length:40},(_,i)=>String(i+1)):annualDates(data.startDate);}
  function completeEntries(req,data,today){return (req===YEAR?Object.keys(data.entries||{}).filter(key=>Number.isFinite(dateValue(key))).sort():keysFor(req,data)).filter(key=>{
    const e=data.entries?.[key];return e?.completed&&entryErrors(req===FORTY?'forty':'year',e,today).length===0&&(req!==YEAR||key<=today);
  });}
  const model={dateValue,addDays,annualDates,chapter,entryErrors,keysFor,completeEntries,YEAR,FORTY,PROGRAM};
  root.SGL531Model=model;
  if(typeof document==='undefined')return;

  const stylesheet=document.createElement('link');stylesheet.rel='stylesheet';stylesheet.href='music-devotion.css?v=5.47';document.head.append(stylesheet);
  let editor=null,leavePending=null,previewUrl='',busyChoice=false,openSequence=0;
  const e=value=>esc(String(value??''));
  const js=value=>e(JSON.stringify(String(value??'')));
  const title=(req,choice='')=>req===FORTY?'Plan 40 días':choice==='reavivados'?'Reavivados por su Palabra':choice==='bible_plan'?'BIBLE PLAN':'Reavivados / Bible Plan / Certificado PDF';
  const target=()=>String(targetAspirant?.()?.id||currentUser?.()?.id||'');
  function context(req,owner=target()){
    if(![YEAR,FORTY].includes(req))throw new Error('Requisito no válido.');
    const actor=currentUser?.(),aspirant=userById(owner);
    if(!actor||!aspirant||!(window.SGL555?SGL555.personalPrograms(aspirant).includes(PROGRAM):v3UserHasProgram(aspirant,PROGRAM))||!(sgl548ViewAll(actor)||actor.role==='regional'&&accessibleAspirants(actor).some(u=>u.id===owner)||['aspirant','pastor_aspirant'].includes(actor.role)&&actor.id===owner||actor.role==='tutor'&&accessibleAspirants(actor).some(u=>u.id===owner)))throw new Error('Este diario no está disponible para tu usuario.');
    const progress=v3EnsureProgramProgress(owner,PROGRAM),sub=v3SubmissionFor(owner,req,PROGRAM),status=progress.status[req]||'pending';
    return {req,owner,actor:actor.id,aspirant,sub,progress,status,editable:['aspirant','pastor_aspirant'].includes(actor.role)&&actor.id===owner&&!['in_review','approved'].includes(status)};
  }
  function dataFor(ctx){
    if(ctx.sub.bibleReadings?.kind===KIND)return clone(ctx.sub.bibleReadings);
    const old=database.bibleYear?.[ctx.owner]?.[PROGRAM];
    return {kind:KIND,version:1,choice:ctx.req===FORTY?'forty':'',startDate:localDate(),entries:{},
      ...(ctx.sub.text||ctx.sub.bibleReadings||old?{legacy:{text:ctx.sub.text||'',readings:clone(ctx.sub.bibleReadings||{}),year:ctx.req===YEAR?clone(old||{}):{}}}:{})};
  }
  function setContext(ctx){v3SetReportContext(PROGRAM,ctx.owner,ctx.req);}
  function assertEditable(ctx){const fresh=context(ctx.req,ctx.owner);if(fresh.actor!==ctx.actor||!fresh.editable)throw new Error('El diario ya no está disponible para edición.');return fresh;}
  function entryText(req,key,item){
    return [req===FORTY?`Día ${key} · ${item.theme||''}`:`Lectura del ${key}`,
      `Leído: ${item.date||''}${item.time?' · '+item.time:''} · ${item.reference||''}`,
      `Versículo favorito: ${item.favoriteVerses||''}`,`Lo que más me impactó: ${item.impactStory||''}`,
      `Aplicación para mi vida: ${item.lifeLesson||''}`,item.aboutGod?`Qué descubrí de Dios: ${item.aboutGod}`:'',
      req===FORTY?`Himno o canto y cómo se relaciona con el tema: ${item.songTheme||[item.song,item.songRelation].filter(Boolean).join(' — ')}`:''].filter(Boolean).join('\n');
  }
  function syncText(ctx,data){
    const certificate=ctx.sub.primaryDocumentId&&String(ctx.sub.submissionMode||ctx.sub.deliveryMode||'').toLowerCase()==='document';
    ctx.sub.bibleReadings=data;ctx.sub.bibleReadingVersion='5.31';
    if(!certificate){ctx.sub.submissionMode='written';ctx.sub.deliveryMode='written';ctx.sub.documentMode=false;}
    if(data.choice==='bible_plan')return;
    const keys=completeEntries(ctx.req,data,localDate());
    ctx.sub.text=`${title(ctx.req,data.choice)}\n${keys.length} registros completados\n\n`+keys.map(key=>entryText(ctx.req,key,data.entries[key])).join('\n\n');
    ctx.sub.richText=plainTextToRichHtml(ctx.sub.text);ctx.sub.completionDate=keys.map(key=>data.entries[key].date).sort().at(-1)||localDate();
    if(!certificate){ctx.sub.submissionMode='written';ctx.sub.deliveryMode='written';ctx.sub.documentMode=false;}
  }
  async function storeData(ctx,data){
    ctx=assertEditable(ctx);setContext(ctx);
    const previous=clone(ctx.sub),status=ctx.progress.status[ctx.req],updated=ctx.progress.updatedAt;
    syncText(ctx,data);ctx.sub.updatedAt=isoNow();ctx.sub.draftOnly=true;ctx.progress.status[ctx.req]='in_progress';ctx.progress.updatedAt=isoNow();
    try{await persist(true,{action:'Avance guardado',entity:ctx.req,ownerId:ctx.owner,programId:PROGRAM,detail:title(ctx.req)});}
    catch(error){Object.keys(ctx.sub).forEach(key=>delete ctx.sub[key]);Object.assign(ctx.sub,previous);ctx.progress.status[ctx.req]=status;ctx.progress.updatedAt=updated;throw error;}
  }
  const btn=(label,code,primary=false,disabled=false)=>`<button type="button" class="btn ${primary?'primary':'ghost'}" onclick="${e(code)}" ${disabled?'disabled':''}>${e(label)}</button>`;
  function shell(ctx,heading,body,foot=''){
    setContext(ctx);if(window.innerWidth<=760)closeMobileMenu?.();showModal(`<div class="modal sgl531-modal"><div class="modal-head"><div><div class="eyebrow">Sistema de Gestión de Música · ${e(ctx.req)}</div><h3>${e(heading)}</h3><p>${e(ctx.aspirant.fullName)} · ${e(STATUS_LABELS[ctx.status]||ctx.status)}</p></div><button class="modal-close" onclick="closeModal()" aria-label="Cerrar">×</button></div><div class="modal-body">${body}</div><div class="modal-foot">${foot||btn('Cerrar','closeModal()')}</div></div>`);
  }
  function legacyHtml(data){
    if(!data.legacy)return '';
    const old=data.legacy.year?.entries||{};
    const text=data.legacy.text||Object.values(old).map(x=>`${x.date||''} · ${x.reference||''}\n${x.reflection||''}`).join('\n\n')||JSON.stringify(data.legacy.readings||{},null,2);
    return `<details class="sgl531-legacy"><summary>Consultar registros de la versión anterior</summary><p>Se conservan como antecedente. Completa los nuevos campos para incorporarlos al diario.</p><pre>${e(text)}</pre></details>`;
  }
  async function open(req,owner=target(),forceChoice=false){
    if(req===FORTY&&v3SubmissionFor(owner,req,PROGRAM)?.primaryDocumentId&&forceChoice!=='readings'){v3SetReportContext(PROGRAM,owner,req);return window.sgl531OpenDocumentRequirement(req);}
    if(!await beforeLeave())return;
    try{
      const seq=++openSequence,ctx=context(req,owner),data=dataFor(ctx);setContext(ctx);
      if(req===YEAR&&(!data.choice||forceChoice!=='continue')){
        const locked=!!data.choice;
        shell(ctx,title(YEAR),`<p>Elige cómo presentar tu lectura anual. Puedes llenar Reavivados, realizar Bible Plan o subir tu certificado PDF.</p><div class="sgl531-options"><article><h3>Reavivados por su Palabra</h3><p>Un capítulo por día, versículo favorito, reflexión y aplicación. Lectura de hoy o por fecha.</p>${btn('Elegir Reavivados',`SGL531.choose('reavivados',${JSON.stringify(owner)})`,true,!ctx.editable&&data.choice!=='reavivados')}</article><article><h3>Bible Plan</h3><p>Lecturas realizadas aparte; validación mediante el botón Concluido.</p>${btn('Elegir Bible Plan',`SGL531.choose('bible_plan',${JSON.stringify(owner)})`,true,!ctx.editable&&data.choice!=='bible_plan')}</article><article><h3>Certificado PDF</h3><p>Presenta el certificado de tu lectura anual y envíalo a revisión.</p>${btn(ctx.editable?'Subir certificado PDF':'Ver certificado PDF',`SGL531.certificate(${JSON.stringify(owner)})`,true,!ctx.editable&&!ctx.sub.primaryDocumentId)}</article></div><p>El PDF se envía a revisión. Tus registros anteriores se conservan.</p>${locked?'<p>Puedes cambiar o quitar la elección sin borrar tus lecturas; el certificado también sirve como entrega para revisión.</p>':''}${locked&&ctx.editable?btn('Quitar elección',`SGL531.clearChoice(${JSON.stringify(owner)})`):''}${legacyHtml(data)}`);return;
      }
      // La entrega por certificado no impide continuar el diario personal.
      if(data.choice==='bible_plan'){
        shell(ctx,'Bible Plan',`<p>Realiza las lecturas por tu cuenta. Administración, Departamental o Secretaría validará el botón Concluido en Certificados.</p><p class="notice">${ready(ctx)?'Lectura validada como concluida. Ya puedes enviar a revisión.':'Pendiente de validación con Concluido.'}</p>`,btn('Cerrar','closeModal()')+btn('Ver certificados',"SGL550.open('certificates')")+(ctx.editable&&ready(ctx)?btn('Enviar a revisión',`SGL531.send('${req}',${JSON.stringify(owner)})`,true):''));return;
      }
      if(req===YEAR)return day(req,owner,localDate());
      const keys=keysFor(req,data),done=completeEntries(req,data,localDate()),drafts=keys.filter(k=>data.entries?.[k]&&!done.includes(k)).length;
      const first=keys.find(k=>!done.includes(k)&&(req===FORTY||k<=localDate()));
      const action=(name,call,primary=false)=>btn(name,`SGL531.${call}('${req}',${JSON.stringify(owner)})`,primary);
      const annual=req===YEAR?`<label class="field-group">Inicio de mi año de lectura<input id="sgl531_start" type="date" class="field" value="${e(data.startDate)}" max="${localDate()}" ${ctx.editable?'':'disabled'}></label>${ctx.editable?btn('Aplicar fecha',`SGL531.start(${JSON.stringify(owner)})`):''}<p>Un capítulo por fecha del calendario. Puedes completar fechas anteriores sin perder los registros guardados. La fecha de lectura real se registra aparte.</p>`:'<p>Registra una lectura por fecha real, hasta completar 40 días. El tema y el pasaje los eliges según tu material de estudio; el himno o canto se relaciona con ese tema.</p>';
      const calendar=keys.map(k=>{const completed=done.includes(k),draft=data.entries?.[k],future=req===YEAR&&k>localDate();return `<button class="sgl531-day ${completed?'done':draft?'draft':''}" ${future?'disabled':''} onclick="SGL531.day('${req}',${js(owner)},${js(k)})"><b>${e(req===FORTY?'Día '+k:k.slice(8,10)+'/'+k.slice(5,7))}</b><span>${e(req===FORTY?draft?.theme||'Adoración':chapter(k,BIBLE_CANON).reference)}</span><small>${future?'Próxima':completed?'Completada':draft?'Borrador':'Pendiente'}</small></button>`;}).join('');
      const picker=req===YEAR?`<div class="actions"><label class="field-group">Ir a una fecha<input class="field" type="date" id="sgl531_jump" min="${e(keys[0])}" max="${e(keys.at(-1)<localDate()?keys.at(-1):localDate())}" value="${e(first||keys[0])}"></label>${btn('Abrir lectura',`SGL531.day('${req}',${JSON.stringify(owner)},document.getElementById('sgl531_jump').value)`)}</div>`:'';
      const review=v3CanReviewRequirement(currentUser(),ctx.aspirant,ctx.status)?btn('Solicitar corrección',"reviewRequirement('needs_changes')")+btn('Marcar revisado',"reviewRequirement('approved')",true):'';
      shell(ctx,title(req),`<div class="sgl531-progress"><strong>${done.length} de ${keys.length} completadas</strong><span>${drafts} borrador${drafts===1?'':'es'}</span><progress max="${keys.length}" value="${done.length}"></progress></div>${ctx.sub.feedback?`<div class="notice warn">${e(ctx.sub.feedback)}</div>`:''}<div class="sgl531-setup">${annual}</div><div class="actions">${first?btn(req===FORTY?'Registrar siguiente día':'Continuar pendiente',`SGL531.day('${req}',${JSON.stringify(owner)},${JSON.stringify(first)})`,true):''}${action('Ver avances','summary')}${req===YEAR?btn('Ver elección',`SGL531.open('${req}',${JSON.stringify(owner)},true)`):''}</div>${picker}<div class="sgl531-calendar">${calendar}</div>${legacyHtml(data)}`,`${btn('Cerrar','closeModal()')}${ctx.editable?btn('Enviar a revisión',`SGL531.send('${req}',${JSON.stringify(owner)})`,true,done.length!==keys.length):''}${review}`);
    }catch(error){toast(error.message);}
  }
  async function choose(choice,owner){
    if(busyChoice||!['reavivados','bible_plan'].includes(choice))return;busyChoice=true;
    try{let ctx=context(YEAR,owner),data=dataFor(ctx);
      if(data.choice===choice&&(!ctx.editable||!ctx.sub.primaryDocumentId))return await open(YEAR,owner,'continue');
      ctx=assertEditable(ctx);
      if(data.choice&&data.choice!==choice){if(!confirm('¿Cambiar de modalidad? Tus lecturas se conservarán.'))return;}
      data.choice=choice;
      const previous=clone(ctx.sub);
      data.choiceUpdatedAt=isoNow();
      try{await storeData(ctx,data);}catch(error){Object.keys(ctx.sub).forEach(k=>delete ctx.sub[k]);Object.assign(ctx.sub,previous);throw error;}
      await open(YEAR,owner,'continue');
    }catch(error){toast(error.message);}finally{busyChoice=false;}
  }
  async function clearChoice(owner=target()){
    if(busyChoice||!await beforeLeave())return;
    busyChoice=true;
    try{const ctx=assertEditable(context(YEAR,owner)),data=dataFor(ctx);
      if(!confirm('¿Quitar la elección? Se conservarán tus lecturas y el certificado adjunto.'))return;
      const previous=clone(ctx.sub);data.choice='';data.choiceUpdatedAt=isoNow();ctx.sub.bibleReadings=data;ctx.sub.updatedAt=isoNow();setContext(ctx);
      try{await persist(true,{action:'Avance guardado',entity:YEAR,ownerId:owner,programId:PROGRAM,detail:'Elección retirada; registros conservados'});}catch(error){Object.keys(ctx.sub).forEach(k=>delete ctx.sub[k]);Object.assign(ctx.sub,previous);throw error;}
      await open(YEAR,owner);
    }catch(error){toast(error.message);}finally{busyChoice=false;}
  }
  async function certificate(owner=target()){
    if(!await beforeLeave())return;
    try{const ctx=context(YEAR,owner);if(!ctx.editable&&!ctx.sub.primaryDocumentId)throw new Error('No hay un certificado PDF disponible.');setContext(ctx);await window.sgl531OpenDocumentRequirement(YEAR);
      if(currentUser()?.id!==ctx.actor||modalOwnerId!==owner||modalRequirementId!==YEAR)return;
      const head=document.querySelector('.sgl5078-document-modal .modal-head');if(head){const heading=head.querySelector('h3');if(heading)heading.textContent='Lectura anual · Certificado PDF';}
    }catch(error){toast(error.message);}
  }
  async function start(owner){try{const seq=++openSequence,ctx=assertEditable(context(YEAR,owner)),data=dataFor(ctx),date=document.getElementById('sgl531_start')?.value;if(!annualDates(date).length||date>localDate())throw new Error('Elige una fecha de inicio válida, hasta hoy.');data.startDate=date;await storeData(ctx,data);if(seq===openSequence&&currentUser()?.id===ctx.actor)await open(YEAR,owner);}catch(error){toast(error.message);}}
  function readForm(state=editor){
    if(!state)return null;
    const form=document.getElementById('sgl531_form');if(!form)return null;
    const entry={};for(const node of form.querySelectorAll('[data-entry]'))entry[node.dataset.entry]=node.value;
    if(state.req===YEAR){entry.reference=chapter(state.key,BIBLE_CANON).reference;entry.aboutGod=state.original?.aboutGod||'';}else if(state.req===FORTY){entry.reference=state.reference||state.original?.reference||'';}
    entry.completed=!!state.original?.completed;return entry;
  }
  const fingerprint=entry=>JSON.stringify(entry);
  function dirty(){return !!editor&&!!readForm()&&fingerprint(readForm())!==editor.saved;}
  const emergencyKey=state=>`sgl531-draft:${state.actor}:${state.owner}:${state.req}:${state.key}`;
  function emergency(){if(!editor||currentUser()?.id!==editor.actor||!dirty())return;try{localStorage.setItem(emergencyKey(editor),JSON.stringify({actor:editor.actor,owner:editor.owner,base:editor.base,entry:readForm(),at:isoNow()}));}catch{}}
  async function day(req,owner,key){
    if(!await beforeLeave())return;
    try{
      const ctx=context(req,owner),data=dataFor(ctx),seq=++openSequence;
      if(req===YEAR&&data.choice!=='reavivados')return open(req,owner);
      if(req===YEAR?(!Number.isFinite(dateValue(key))||key>localDate()):!keysFor(req,data).includes(String(key)))throw new Error('Selecciona una fecha del período, hasta hoy.');
      key=String(key);let entry=clone(data.entries?.[key]||{date:localDate(),favoriteVerses:'',impactStory:'',lifeLesson:'',aboutGod:'',...(req===FORTY?{theme:'',reference:'',song:'',songRelation:'',time:''}:{reference:chapter(key,BIBLE_CANON).reference}),completed:false});
      const state={...ctx,key,base:ctx.sub.updatedAt||'',original:clone(entry),saved:'',saving:null};let recovered=false;
      try{const draft=JSON.parse(localStorage.getItem(emergencyKey(state))||'null');if(ctx.editable&&draft?.actor===ctx.actor&&draft.owner===ctx.owner&&draft.base===state.base){entry=draft.entry;recovered=true;}}catch{}
      const bible=null; // The passage is loaded only when the reader is opened.
      if(seq!==openSequence||currentUser()?.id!==ctx.actor)return;
      const field=(key,label,rows=3,type='textarea',optional=false)=>`<label class="field-group sgl531-field ${({favoriteVerses:"sgl519-prompt-card favorite",impactStory:"sgl519-prompt-card impact",lifeLesson:"sgl519-prompt-card lesson",aboutGod:"sgl519-prompt-card lesson",song:"sgl519-prompt-card favorite",songRelation:"sgl519-prompt-card impact"})[key]||""}">${["favoriteVerses","impactStory","lifeLesson"].includes(key)?`<span class="sgl519-prompt-icon" aria-hidden="true">${({favoriteVerses:"V",impactStory:"M",lifeLesson:"A"})[key]}</span>`:""}<span>${e(label)}${optional?' (opcional)':''}</span>${type==='textarea'?`<textarea class="field" data-entry="${key}" ${key==='favoriteVerses'?'id="reading_verse"':''} rows="${rows}" ${ctx.editable?'':'readonly'}>${e(entry[key]||'')}</textarea>`:`<input class="field" data-entry="${key}" type="${type}" value="${e(entry[key]||'')}" ${type==='date'?`max="${localDate()}"`:''} ${ctx.editable?'':'readonly'}>`}</label>`;
      const form=`<form id="sgl531_form" onsubmit="event.preventDefault()"><div class="sgl531-two">${field('date','Fecha real de lectura',1,'date')}${req===FORTY?field('time','Hora',1,'time',true):`<div class="notice"><b>${e(chapter(key,BIBLE_CANON).reference)}</b><p>Lectura programada: ${e(key)}</p></div>`}</div>${req===FORTY?field('theme','Tema de adoración',2)+field('reference','Pasaje bíblico',1,'text'):''}${bible?`<details open class="sgl519-reading-scripture"><summary><b>${e(bible.name)} · ${e(chapter(key,BIBLE_CANON).reference)}</b></summary><div id="bible_copy">${bible.html}</div></details>`:''}<div class="sgl519-reflection-grid">${field('favoriteVerses','Versículo o versículos favoritos',2)}${field('impactStory','¿Qué mensaje me impactó?',4)}${field('lifeLesson','¿Cómo lo aplicaré en mi vida?',4)}${field('aboutGod','¿Qué descubrí de Dios?',3,'textarea',true)}${req===FORTY?field('song','Himno o canto',2)+field('songRelation','¿Cómo se relaciona con el tema?',3):''}</div></form><p id="sgl531_save_state" role="status" aria-live="polite">${ctx.editable?recovered?'Borrador recuperado de este usuario.':'':'Consulta del diario. Solo su propietario puede editarlo.'}</p><p id="sgl531_error" class="notice warn" hidden></p>`;
      if([YEAR,FORTY].includes(req))renderYearDay(ctx,data,key,entry,recovered);else shell(ctx,req===FORTY?'Día '+key:`Lectura · ${key}`,form,btn('Volver al diario',`SGL531.open('${req}',${JSON.stringify(owner)})`)+(ctx.editable?btn('Guardar borrador','SGL531.save(false)')+btn('Guardar lectura','SGL531.save(true)',true)+btn('Guardar y siguiente','SGL531.save(true,true)'):''));
      editor=state;if(req===FORTY)editor.reference=ctx.reference;editor.saved=fingerprint(readForm());if(recovered)editor.saved='';
      document.getElementById('sgl531_form').addEventListener('input',()=>{emergency();const status=document.getElementById('sgl531_save_state');if(status){status.textContent='';status.hidden=true;}});
      document.querySelectorAll('#bible_copy .verse').forEach(node=>{if(String(entry.favoriteVerses).split(';').map(x=>x.trim()).includes(node.dataset.reference))node.classList.add('selected');if(!ctx.editable)node.disabled=true;});
    }catch(error){toast(error.message);}
  }
  function renderYearDay(ctx,data,key,entry,recovered){
    const forty=ctx.req===FORTY,done=completeEntries(ctx.req,data,localDate()),total=keysFor(ctx.req,data).length,reference=forty?(entry.reference||'Salmos 1'):chapter(key,BIBLE_CANON).reference;
    if(forty){entry.songTheme ||= [entry.song,entry.songRelation].filter(Boolean).join(' — ');ctx.reference=reference;}
    const card=(name,label,hint,kind,mark)=>`<label class="sgl519-prompt-card ${kind}"><span class="sgl519-prompt-icon" aria-hidden="true">${mark}</span><b>${e(label)}</b><small>${e(hint)}</small><textarea class="field" data-entry="${name}" ${name==='favoriteVerses'?'id="reading_verse"':''} rows="${name==='favoriteVerses'?2:4}" ${ctx.editable?'':'readonly'}>${e(entry[name]||'')}</textarea></label>`;
    const review=v3CanReviewRequirement(currentUser(),ctx.aspirant,ctx.status)?btn('Solicitar corrección',"reviewRequirement('needs_changes')")+btn('Marcar revisado',"reviewRequirement('approved')",true):'';
    const body=`<section class="sgl533-hero"><span class="eyebrow">${forty?"Plan 40 días / 40 madrugadas":"Reavivados por su Palabra"}</span><h2>${forty?"Día "+key+" de 40":"Cada día, un capítulo…"}</h2><p>Cada día, una nueva promesa.</p><div class="sgl533-progress-row"><span><strong data-sgl533-count>${done.length}</strong> lecturas completadas</span>${btn('Ver avances',`SGL531.summary('${ctx.req}',${JSON.stringify(ctx.owner)})`,false,!done.length)}</div>${forty?`<progress class="sgl550-reading-progress" max="40" value="${done.length}"></progress>`:''}</section>
      ${ctx.sub.feedback?`<p class="notice warn">${e(ctx.sub.feedback)}</p>`:''}
      ${forty?`<div class="sgl533-date-choice"><label class="field-group">Texto bíblico · Libro<select class="field" id="sgl550_book" ${ctx.editable?'':'disabled'} onchange="SGL531.selectChapter()">${BIBLE_CANON.map(([code,name,count])=>`<option value="${code}" ${reference.startsWith(name+" ")?"selected":""}>${e(name)}</option>`).join("")}</select></label><label class="field-group">Capítulo<input class="field" id="sgl550_chapter" ${ctx.editable?'':'disabled'} type="number" min="1" value="${Number(reference.match(/\d+$/)?.[0])||1}" onchange="SGL531.selectChapter()"></label></div>`:`<div class="sgl533-date-choice">${btn('Lectura de hoy',`SGL531.pick(${JSON.stringify(ctx.owner)},localDate(),true)`,true)}<label class="field-group">Escoge una fecha para leer<input id="sgl533_read_date" class="field" type="date" max="${localDate()}" value="${e(key)}" onchange="SGL531.pick(${js(ctx.owner)},this.value)"></label></div>`}
      <section class="sgl533-passage"><div><small>Lectura elegida</small><h3>${e(reference)}</h3><span>${e(formatDate(forty?entry.date:key))}</span></div><button type="button" id="sgl533_read_passage" class="btn primary" onclick="SGL531.readPassage()" aria-expanded="false" aria-controls="sgl533_reader">Abrir capítulo · Leer pasaje</button></section>
      <dialog id="sgl533_reader" class="sgl534-reading-dialog" aria-labelledby="sgl534_chapter"><header><div><h2 id="sgl534_chapter">${e(reference)}</h2><p id="sgl533_bible_name"></p></div><button class="btn ghost" type="button" onclick="SGL531.closePassage()" aria-label="Cerrar pasaje">Cerrar pasaje ×</button></header><p class="sgl534-reader-hint">Selecciona los versículos que deseas guardar como favoritos.</p><div id="bible_copy" class="bible-text"></div><footer><button class="btn primary" type="button" onclick="SGL531.closePassage()">Volver a mi reflexión</button></footer></dialog>
      <form id="sgl531_form" onsubmit="event.preventDefault();SGL531.save(true)">${forty?`<label class="field-group">Tema de adoración<input class="field" data-entry="theme" value="${e(entry.theme||"")}" ${ctx.editable?"":"readonly"}></label><input type="hidden" data-entry="reference" id="sgl550_reference" value="${e(reference)}">`:""}<section class="sgl519-reflection-grid sgl533-reflections">${card('favoriteVerses','Versículo favorito','Selecciona un versículo al leer el pasaje o escríbelo aquí.','favorite','V')}${card('impactStory','Mensaje e historia que me impactó','¿Qué llamó tu atención en este capítulo?','impact','M')}${card('lifeLesson','Cómo aplicaré en mi vida','Escribe una acción para poner en práctica.','lesson','A')}${forty?card('songTheme','Himno o canto y cómo se relaciona con el tema','Escribe el canto y su relación con la lectura.','impact','♫'):''}</section><label class="field-group sgl519-date-field">Fecha de lectura<input class="field" type="date" data-entry="date" value="${e(entry.date||localDate())}" max="${localDate()}" ${ctx.editable?'':'readonly'}></label></form>
      <p id="sgl531_save_state" role="status" aria-live="polite" ${recovered?'':'hidden'}>${recovered?'Borrador recuperado.':''}</p><p id="sgl531_error" class="notice warn" hidden></p>`;
    shell(ctx,title(ctx.req,'reavivados'),body,btn('Cerrar','closeModal()')+(ctx.editable?btn('Guardar lectura','SGL531.save(true)',true):'')+(ctx.editable&&ready(ctx)?btn('Enviar a revisión',`SGL531.send('${ctx.req}',${JSON.stringify(ctx.owner)})`):'')+review);
    document.querySelector('.sgl531-modal').classList.add('sgl533-reavivados');
  }
  function selectChapter(){
    if(!editor||editor.req!==FORTY)return;
    const book=BIBLE_CANON.find(b=>b[0]===document.getElementById('sgl550_book')?.value),input=document.getElementById('sgl550_chapter');
    if(!book||!input)return;const n=Math.max(1,Math.min(book[2],Number(input.value)||1));input.value=n;input.max=book[2];
    const reference=book[1]+' '+n;
    if(!editor.editable){input.value=Number(editor.reference.match(/\d+$/)?.[0])||1;return;}
    editor.reference=reference;document.getElementById('sgl550_reference').value=reference;document.querySelector('.sgl533-passage h3').textContent=reference;document.getElementById('sgl534_chapter').textContent=reference;
    document.getElementById('reading_verse').value='';document.getElementById('sgl531_form').dispatchEvent(new Event('input',{bubbles:true}));
  }
  let picking=false;
  async function pick(owner,key,openReader=false){
    if(picking)return;const prior=editor,previous=editor?.key;
    if(!Number.isFinite(dateValue(key))||key>localDate()){toast('Escoge una fecha válida, hasta hoy.');const input=document.getElementById('sgl533_read_date');if(input)input.value=previous||localDate();return;}
    picking=true;
    if(!await beforeLeave()){picking=false;const input=document.getElementById('sgl533_read_date');if(input)input.value=previous||localDate();return;}
    const form=document.getElementById('sgl531_form');if(form)form.inert=true;
    try{
      const ctx=context(YEAR,owner);
      if(currentUser()?.id!==ctx.actor)return;
      await day(YEAR,owner,key);
      if(openReader&&editor?.key===key&&editor?.owner===owner)await readPassage();
    }catch(error){if(prior&&currentUser()?.id===prior.actor&&form?.isConnected)editor=prior;toast(error.message);const input=document.getElementById('sgl533_read_date');if(input)input.value=previous||localDate();}finally{picking=false;if(form?.isConnected)form.inert=false;}
  }
  function closePassage(){
    const panel=document.getElementById('sgl533_reader');
    if(panel?.open)panel.close();
  }
  async function readPassage(){
    const state=editor;if(!state||![YEAR,FORTY].includes(state.req)||currentUser()?.id!==state.actor)return;
    const panel=document.getElementById('sgl533_reader'),button=document.getElementById('sgl533_read_passage');if(!panel||!button)return;
    if(panel.open){closePassage();return;}
    button.disabled=true;
    try{
      const bible=await bibleTextForPassage(state.req===FORTY?state.reference:chapter(state.key,BIBLE_CANON).reference);
      if(editor!==state||currentUser()?.id!==state.actor||!panel.isConnected)return;
      document.getElementById('sgl533_bible_name').textContent=bible.name;
      const copy=document.getElementById('bible_copy'),source=document.createElement('div');source.innerHTML=bible.html;
      const input=document.getElementById('reading_verse'),selected=()=>String(input.value).split(';').map(x=>x.trim()).filter(Boolean);
      copy.replaceChildren();
      for(const row of source.querySelectorAll('.verse')){
        // Only verse rows become selectable: chapter titles and headings stay outside.
        const number=row.querySelector('sup')?.textContent?.trim()||'';
        const ref=row.dataset.reference||(/^[0-9]+:[0-9]+$/.test(number)?(state.req===FORTY?state.reference.replace(/ \d+$/,''):chapter(state.key,BIBLE_CANON).name)+' '+number:'');
        if(!ref)continue;
        const verse=document.createElement('button');verse.type='button';verse.className='verse';verse.dataset.reference=ref;
        const sup=document.createElement('sup');sup.textContent=number;row.querySelector('sup')?.remove();
        verse.append(sup,document.createTextNode(row.textContent.trim()));
        verse.classList.toggle('selected',selected().includes(ref));verse.setAttribute('aria-pressed',String(verse.classList.contains('selected')));verse.disabled=!state.editable;
        verse.onclick=()=>{if(editor!==state||currentUser()?.id!==state.actor)return;try{if(!context(state.req,state.owner).editable)return;}catch{return;}const refs=selected(),index=refs.indexOf(ref);if(index<0)refs.push(ref);else refs.splice(index,1);input.value=refs.join('; ');verse.classList.toggle('selected',index<0);verse.setAttribute('aria-pressed',String(index<0));input.dispatchEvent(new Event('input',{bubbles:true}));};
        copy.append(verse);
      }
      if(!copy.children.length){const note=document.createElement('p');note.textContent=source.textContent.trim()||'No hay versículos disponibles para este pasaje.';copy.append(note);}
      panel.onkeydown=event=>{if(event.key==='Escape')event.stopPropagation();};
      panel.onclose=()=>{button.setAttribute('aria-expanded','false');button.textContent='Abrir capítulo · Leer pasaje';};
      panel.showModal();copy.scrollTop=0;button.setAttribute('aria-expanded','true');button.textContent='Cerrar pasaje';
    }catch(error){if(editor===state)toast('No se pudo abrir el pasaje: '+error.message);}finally{if(button.isConnected)button.disabled=false;}
  }

  async function save(completed=false,next=false){
    const state=editor;if(!state)return false;if(state.saving)return state.saving;
    state.saving=(async()=>{try{
      const ctx=assertEditable(state),entry=readForm(state);if(!entry)return false;
      const errors=entryErrors(state.req===FORTY?'forty':'year',entry,localDate());
      if(completed&&errors.length){const box=document.getElementById('sgl531_error');box.hidden=false;box.textContent=errors.join(' ');return false;}
      const fingerprintAtStart=fingerprint(entry),data=dataFor(ctx);
      entry.completed=(completed||state.original.completed)&&errors.length===0;entry.updatedAt=isoNow();
      if(state.req===FORTY&&entry.completed&&Object.entries(data.entries||{}).some(([key,item])=>key!==state.key&&item.completed&&item.date===entry.date)){document.getElementById('sgl531_error').hidden=false;document.getElementById('sgl531_error').textContent='Esta fecha ya tiene una madrugada registrada. Indica la fecha real de esta madrugada.';return false;}
      data.entries||={};data.entries[state.key]=entry;await storeData(ctx,data);
      if(currentUser()?.id!==state.actor||editor!==state)return true;
      state.base=context(state.req,state.owner).sub.updatedAt||'';
      state.original=clone(entry);
      // No dar por guardado lo escrito mientras persistía la instantánea anterior.
      state.saved=fingerprintAtStart;
      const current=readForm();if(current&&JSON.stringify({...current,completed:entry.completed})===JSON.stringify({...JSON.parse(fingerprintAtStart),completed:entry.completed}))state.saved=fingerprint(current);
      try{localStorage.removeItem(emergencyKey(state));}catch{}if(dirty())emergency();
      const savedStatus=document.getElementById('sgl531_save_state');if(savedStatus){savedStatus.textContent=completed&&!dirty()?'Lectura guardada.':'';savedStatus.hidden=!savedStatus.textContent;}
      if(state.req===YEAR){const total=completeEntries(YEAR,data,localDate()).length;document.querySelector('[data-sgl533-count]')?.replaceChildren(document.createTextNode(String(total)));const meter=document.querySelector('.sgl533-hero progress');if(meter)meter.value=total;const advances=document.querySelector('.sgl533-progress-row button');if(advances)advances.disabled=!total;
        const foot=document.querySelector('.sgl533-reavivados .modal-foot'),send=foot?.querySelector('[onclick*="SGL531.send"]');
        if(total>=keysFor(YEAR,data).length&&!send&&foot)foot.insertAdjacentHTML('beforeend',btn('Enviar a revisión',`SGL531.send('${ctx.req}',${JSON.stringify(state.owner)})`));
        else if(total<keysFor(YEAR,data).length)send?.remove();
      }

      document.getElementById('sgl531_error').hidden=true;
      if(completed&&next&&!dirty()){const keys=keysFor(state.req,data),done=completeEntries(state.req,data,localDate()),following=keys.slice(keys.indexOf(state.key)+1).find(k=>!done.includes(k)&&(state.req===FORTY||k<=localDate()))||keys.find(k=>!done.includes(k)&&(state.req===FORTY||k<=localDate()));editor=null;following?await day(state.req,state.owner,following):await open(state.req,state.owner);}
      return true;
    }catch(error){emergency();const status=document.getElementById('sgl531_save_state');if(status){status.hidden=false;status.textContent='No se pudo guardar. Tu borrador permanece en este dispositivo. '+error.message;}return false;}})();
    const pending=state.saving;try{return await pending;}finally{if(state.saving===pending)state.saving=null;}
  }
  async function beforeLeave(){
    if(leavePending)return leavePending;
    if(!editor)return true;
    if(currentUser()?.id!==editor.actor){editor=null;return true;}
    if(editor.saving)await editor.saving;
    if(!dirty()){editor=null;return true;}
    leavePending=new Promise(resolve=>{
      const layer=document.createElement('div');layer.className='sgl531-leave';layer.innerHTML=`<div role="alertdialog" aria-modal="true" aria-labelledby="sgl531_leave_title"><h3 id="sgl531_leave_title">¿Guardar los cambios?</h3><p>Hay cambios posteriores al último guardado. Cancelar te mantiene en el diario.</p><div class="actions"><button class="btn primary" data-choice="save">Guardar</button><button class="btn ghost" data-choice="discard">No guardar</button><button class="btn ghost" data-choice="cancel">Cancelar</button></div></div>`;document.body.append(layer);
      const modal=document.querySelector('.modal-bg');if(modal)modal.inert=true;
      layer.querySelector('[data-choice=cancel]').focus();
      layer.addEventListener('keydown',ev=>{if(ev.key==='Escape'){ev.stopPropagation();layer.querySelector('[data-choice=cancel]').click();}if(ev.key==='Tab'){ev.preventDefault();const buttons=[...layer.querySelectorAll('button')];buttons[(buttons.indexOf(document.activeElement)+(ev.shiftKey?2:1))%3].focus();}});
      layer.addEventListener('click',async event=>{const choice=event.target.closest('[data-choice]')?.dataset.choice;if(!choice)return;layer.querySelectorAll('button').forEach(b=>b.disabled=true);let ok=choice==='discard';if(choice==='save')ok=await save(false);if(ok){try{localStorage.removeItem(emergencyKey(editor));}catch{}editor=null;}layer.remove();if(modal)modal.inert=false;leavePending=null;resolve(ok);});
    });return leavePending;
  }
  function ready(ctx){if([YEAR,FORTY].includes(ctx.req)&&ctx.sub.primaryDocumentId&&String(ctx.sub.submissionMode||ctx.sub.deliveryMode||'').toLowerCase()==='document')return true;const data=dataFor(ctx);if(ctx.req===YEAR)return ['reavivados','bible_plan'].includes(data.choice)&&!!database.music550?.completions?.[ctx.owner]?.[data.choice]?.concluded;return !!data.choice&&completeEntries(ctx.req,data,localDate()).length>=keysFor(ctx.req,data).length;}
  async function send(req,owner=target()){
    if([YEAR,FORTY].includes(req)&&v3SubmissionFor(owner,req,PROGRAM)?.primaryDocumentId&&String(v3SubmissionFor(owner,req,PROGRAM)?.submissionMode||v3SubmissionFor(owner,req,PROGRAM)?.deliveryMode||'').toLowerCase()==='document'){v3SetReportContext(PROGRAM,owner,req);return window.sgl5078SendDocumentReview(req);}
    if(!await beforeLeave())return;
    try{const ctx=assertEditable(context(req,owner));if(!ready(ctx))throw new Error(req===YEAR?'Una persona autorizada debe validar Concluido en la modalidad elegida antes de enviar.':'Completa las 40 madrugadas antes de enviar.');
      if(req===YEAR){const d=dataFor(ctx),check=database.music550.completions[owner][d.choice];
        const validation=`Lectura anual validada: ${d.choice}\nValidada por: ${userById(check.validatedBy)?.fullName||check.validatedBy}\nFecha de conclusión: ${check.at}`;
        if(!String(ctx.sub.text||'').includes('Lectura anual validada:'))ctx.sub.text=[ctx.sub.text,validation].filter(Boolean).join('\n\n');
        ctx.sub.richText=plainTextToRichHtml(ctx.sub.text);ctx.sub.submissionMode='written';ctx.sub.deliveryMode='written';ctx.sub.documentMode=false;
        ctx.sub.completionDate ||= String(check.at||'').slice(0,10)||localDate();
      }
      setContext(ctx);await baseSend(req);if(context(req,owner).status==='in_review')await open(req,owner);
    }catch(error){toast(error.message);}
  }

  // Informes acumulativos: se paginan todas las líneas, incluidas las continuaciones.
  function report(ctx,from='',to=''){
    const data=dataFor(ctx),keys=completeEntries(ctx.req,data,localDate()).filter(k=>(!from||k>=from)&&(!to||k<=to));
    if(!keys.length)throw new Error('Todavía no hay lecturas completadas en este período.');
    return {ctx,data,keys,title:title(ctx.req,data.choice),entries:keys.map(key=>({key,item:data.entries[key]}))};
  }
  function wrap(text,width,size=9,font='AR'){
    const lines=[];for(const paragraph of String(text||'').replace(/\r/g,'').split('\n')){
      let line='';for(const word of paragraph.split(/\s+/)){if(!word)continue;const candidate=line?line+' '+word:word;if(measureReportText(candidate,size,font)<=width){line=candidate;continue;}if(line){lines.push(line);line='';}let part='';for(const char of word){if(part&&measureReportText(part+char,size,font)>width){lines.push(part);part='';}part+=char;}line=part;}lines.push(line);
    }return lines;
  }
  const CARD_COLORS=['FDF2F8','FFFAEB','ECFDF5','EFF6FF'];
  function coloredPages(value){
    const longest=Math.max(...value.entries.map(({item})=>String(item.favoriteVerses||'').length+String(item.impactStory||'').length+String(item.lifeLesson||'').length+String(item.aboutGod||'').length));
    const columns=longest<=440?3:longest<=1200?2:1,width=(470-(columns-1)*8)/columns,size=columns===3?7:8,lineHeight=size+2;
    const pages=[];let page={blocks:[],rows:[],used:0,columns,width,size,lineHeight},row=[];
    const pushPage=()=>{if(page.blocks.length)pages.push(page);page={blocks:[],rows:[],used:0,columns,width,size,lineHeight};};
    const flush=()=>{if(!row.length)return;const height=Math.max(...row.map(b=>b.height));if(page.used+height>600||page.blocks.length+row.length>(columns===3?9:columns===2?8:6))pushPage();const y=page.used;row.forEach((block,i)=>{block.x=i*(width+8);block.y=y;page.blocks.push(block);});page.rows.push(row);page.used+=height+8;row=[];};
    for(const [index,{key,item}] of value.entries.entries()){
      const heading=`${String(index+1).padStart(2,'0')}  ${item.reference||key}`;
      const headLines=wrap(heading,width-14,size+1,'AB');
      const fields=[['Versículos favoritos',item.favoriteVerses],['Me impactó',item.impactStory],['Para mi vida',item.lifeLesson],...(item.aboutGod?[['Qué descubrí de Dios',item.aboutGod]]:[])];
      const lines=fields.flatMap(([label,text],color)=>[...wrap(label.toUpperCase(),width-20,size,'AB').map(text=>({text,bold:true,color})),...wrap(text,width-20,size).map(text=>({text,bold:false,color}))]);
      const headerHeight=14+headLines.length*(size+3)+lineHeight;
      // Bound each card to the printable page; long entries continue without clipping.
      for(let offset=0;offset<lines.length;){
        const continued=offset>0;if(continued)flush();
        const blockHeader=headerHeight+(continued?size+3:0);
        const maxLines=Math.max(1,Math.floor((600-blockHeader-8)/lineHeight));
        const part=lines.slice(offset,offset+maxLines),block={heading,headLines:continued?[...headLines,'(continuación)']:headLines,date:item.date+(item.time?' · '+item.time:''),lines:part,headerHeight:blockHeader};
        block.height=block.headerHeight+part.length*lineHeight+8;offset+=part.length;
        row.push(block);if(row.length===columns)flush();
      }
    }
    flush();pushPage();return pages;
  }
  function cardGroups(lines){const groups=[];for(const line of lines){if(groups.at(-1)?.color!==line.color)groups.push({color:line.color,lines:[]});groups.at(-1).lines.push(line);}return groups;}
  function coloredCardHtml(block,page){return `<section class="sgl535-card" style="left:${block.x}pt;top:${block.y}pt;width:${page.width}pt;height:${block.height}pt;font-size:${page.size}pt;line-height:${page.lineHeight}pt"><header style="height:${block.headerHeight}pt;box-sizing:border-box">${block.headLines.map(line=>`<b>${e(line)}</b>`).join('')}<small>${e(formatDate(block.date))}</small></header>${cardGroups(block.lines).map(group=>`<div style="background:#${CARD_COLORS[group.color%4]}">${group.lines.map(line=>`<span style="font-weight:${line.bold?'700':'400'}">${e(line.text)||'&nbsp;'}</span>`).join('')}</div>`).join('')}</section>`;}
  async function reportHeader(value){
    const club=clubFor(value.ctx.aspirant),official=await v3LoadProgramLogo(PROGRAM);
    return {official,club:club?.logo||'',mission:reportMissionName({aspirant:value.ctx.aspirant}),title:v3Program(PROGRAM)?.report_title||'INFORME LÍDER DE MÚSICA',name:value.ctx.aspirant.fullName};
  }
  async function coloredPdf(value){
    const prepared=coloredPages(value),pages=[],header=await reportHeader(value);
    const logo=await jpegResource(header.official,480,true),clubLogo=header.club?await jpegResource(header.club,480,true):null;
    const rgb=hex=>hex.match(/../g).map(x=>(parseInt(x,16)/255).toFixed(3)).join(' ');
    for(const [index,content] of prepared.entries()){
      const page={w:612,h:792,commands:[],images:[]};pages.push(page);
      for(const [img,x] of [[logo,56],[clubLogo,508]])if(img){const ratio=Math.min(48/img.w,43/img.h);pdfImage(page,img,x,23,img.w*ratio,img.h*ratio);}
      let y=20;
      for(const [text,size,font] of [[header.mission,8,'AB'],[header.title,12,'AB'],[header.name,10,'AR']]){
        for(const line of wrap(text,390,size,font)){pdfText(page,line,306-measureReportText(line,size,font)/2,y,size,font,false,font==='AB'&&size===12?'0.122 0.451 0.620':'0 0 0');y+=size+3;}
      }
      page.commands.push('0.122 0.451 0.620 RG 1.5 w 85 709 m 555 709 l S');
      pdfText(page,value.title,85,94,10,'AB');pdfText(page,`${value.keys.length} lecturas · ${value.ctx.req}`,85,109,8,'AR');
      for(const block of content.blocks){
        const x=85+block.x,top=132+block.y,w=content.width;
        page.commands.push(`0.84 0.88 0.91 RG 0.6 w ${x} ${792-top-block.height} ${w} ${block.height} re S`);
        let y=top+7;for(const line of block.headLines){pdfText(page,line,x+7,y,content.size+1,'AB');y+=content.size+3;}
        pdfText(page,block.date,x+7,y,content.size-1,'AR',false,'0.35 0.40 0.43');y=top+block.headerHeight;
        for(const group of cardGroups(block.lines)){
          const h=group.lines.length*content.lineHeight;
          page.commands.push(`${rgb(CARD_COLORS[group.color%4])} rg ${x+5} ${792-y-h} ${w-10} ${h} re f`);
          for(const line of group.lines){pdfText(page,line.text,x+10,y,content.size,line.bold?'AB':'AR');y+=content.lineHeight;}
        }
      }
      const info=typeof reportInstitutionInfo==='function'?reportInstitutionInfo({aspirant:value.ctx.aspirant}):{};
      pdfText(page,'IGLESIA: '+String(info.church||''),85,767,7,'AR');
      pdfText(page,'DISTRITO: '+String(info.district||''),280,767,7,'AR');
      pdfText(page,'REGIÓN: '+String(info.region||''),510,767,7,'AR');
      pdfText(page,`${index+1} / ${prepared.length}`,535,746,7,'AR');
    }
    return buildVectorPdf(pages);
  }
  function coloredWordTable(page,p){
    const cellWidth=Math.round((470/page.columns)*20);
    return `<w:tbl><w:tblPr><w:tblW w:w="9400" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="70" w:type="dxa"/><w:bottom w:w="70" w:type="dxa"/><w:left w:w="80" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>${Array.from({length:page.columns},()=>`<w:gridCol w:w="${cellWidth}"/>`).join('')}</w:tblGrid>${page.rows.map(row=>`<w:tr><w:trPr><w:cantSplit/></w:trPr>${Array.from({length:page.columns},(_,i)=>{const block=row[i];return `<w:tc><w:tcPr><w:tcW w:w="${cellWidth}" w:type="dxa"/><w:vAlign w:val="top"/></w:tcPr>${block?block.headLines.map(line=>p(line,true,page.size+1)).join('')+p(block.date,false,page.size-1)+block.lines.map(line=>p(line.text,line.bold,page.size).replace('<w:pPr>',`<w:pPr><w:shd w:val="clear" w:fill="${CARD_COLORS[line.color%4]}"/>`)).join(''):p('')}</w:tc>`;}).join('')}</w:tr>`).join('')}</w:tbl>`;
  }

  function reportPages(value){
    if(value.ctx.req===YEAR)return coloredPages(value);
    const pages=[];let page={blocks:[],used:0};pages.push(page);
    for(const {key,item} of value.entries){
      const heading=value.ctx.req===FORTY?`Día ${key}`:`${key} · ${item.reference}`;
      const fields=[['Lectura realizada',item.date+(item.time?' · '+item.time:'')],...(value.ctx.req===FORTY?[['Tema de adoración',item.theme],['Pasaje',item.reference]]:[]),['Versículos favoritos',item.favoriteVerses],['Mensaje que me impactó',item.impactStory],['Aplicación para mi vida',item.lifeLesson],...(item.aboutGod?[['Qué descubrí de Dios',item.aboutGod]]:[]),...(value.ctx.req===FORTY?[['Himno o canto y cómo se relaciona con el tema',item.songTheme||[item.song,item.songRelation].filter(Boolean).join(' — ')]]:[])];
      const lines=fields.flatMap(([label,text])=>[...wrap(label+':',504,9,'AB').map(text=>({text,bold:true})),...wrap(text,504).map(text=>({text,bold:false}))]);
      let offset=0;while(offset<lines.length){
        const max=Math.floor((588-page.used-38)/12);if(max<4){page={blocks:[],used:0};pages.push(page);continue;}
        const part=lines.slice(offset,offset+max),block={heading:heading+(offset?' (continuación)':''),lines:part,height:34+part.length*12};page.blocks.push(block);page.used+=block.height+8;offset+=part.length;
        if(offset<lines.length){page={blocks:[],used:0};pages.push(page);}
      }
    }return pages;
  }
  async function buildPdf(value){
    if(value.ctx.req===YEAR)return coloredPdf(value);
    const prepared=reportPages(value),pages=[];let logo=null;try{logo=await jpegResource(await v3LoadProgramLogo(PROGRAM),150,false);}catch{}
    for(const [index,content] of prepared.entries()){
      const page={w:612,h:792,commands:[],images:[]};pages.push(page);
      if(logo){const ratio=Math.min(48/logo.w,48/logo.h);pdfImage(page,logo,40,28,logo.w*ratio,logo.h*ratio);}
      pdfText(page,'SISTEMA DE GESTIÓN DE MÚSICA',102,30,10,'AB',false,'0.45 0.30 0.08');pdfText(page,value.title,102,46,15,'AB');
      wrap(value.ctx.aspirant.fullName,480,10).forEach((line,i)=>pdfText(page,line,40,83+i*12,10,'AB'));
      pdfText(page,`${value.ctx.req} · ${value.keys.length} lecturas · ${value.keys[0]} a ${value.keys.at(-1)}`,40,109,8,'AR');
      let top=135;for(const block of content.blocks){page.commands.push(`0.89 0.85 0.76 RG 0.5 w 40 ${792-top-block.height} 532 ${block.height} re S`);wrap(block.heading,504,10,'AB').forEach((line,i)=>pdfText(page,line,52,top+7+i*11,10,'AB'));let y=top+24;for(const line of block.lines){pdfText(page,line.text,52,y,9,line.bold?'AB':'AR');y+=12;}top+=block.height+8;}
      pdfText(page,`SGL 5.47 · ${index+1} / ${prepared.length} · Registro acumulativo`,40,755,8,'AR',false,'0.4 0.4 0.4');
    }return buildVectorPdf(pages);
  }
  async function buildWord(value){
    const p=(text,bold=false,size=9)=>`<w:p><w:pPr><w:spacing w:after="0" w:line="${value.ctx.req===YEAR?(size+2)*20:Math.max(240,size*24)}" w:lineRule="exact"/></w:pPr>${sgl5063RunXml(text,{font:'Arial',size,bold})}</w:p>`;
    let logo=null;try{logo=await sgl5063PrepareDocxImage(await v3LoadProgramLogo(PROGRAM),{maxPx:180,quality:.9});}catch{}
    const media=logo?[{name:'word/media/music.jpg',data:logo.bytes}]:[];
    const logoRel=logo?'<Relationship Id="logo" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/music.jpg"/>':'';
    const pages=reportPages(value),body=[];
    for(const [index,page] of pages.entries()){
      if(index)body.push('<w:p><w:r><w:br w:type="page"/></w:r></w:p>');
      const heading=p('SISTEMA DE GESTIÓN DE MÚSICA',true,10)+p(value.title,true,14)+p(value.ctx.aspirant.fullName,true,10);
      if(logo){const size=sgl5063ContainEmu(logo,42*12700,42*12700);body.push(`<w:tbl><w:tblPr><w:tblW w:w="10640" w:type="dxa"/><w:tblLayout w:type="fixed"/></w:tblPr><w:tblGrid><w:gridCol w:w="1200"/><w:gridCol w:w="9440"/></w:tblGrid><w:tr><w:tc><w:tcPr><w:tcW w:w="1200" w:type="dxa"/></w:tcPr><w:p>${sgl5063DrawingXml('logo',size.cx,size.cy,index+1,'Sistema de Gestión de Música')}</w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="9440" w:type="dxa"/></w:tcPr>${heading}</w:tc></w:tr></w:tbl>`);}else body.push(heading);
      body.push(p(`${value.ctx.req} · ${value.keys.length} lecturas · ${value.keys[0]} a ${value.keys.at(-1)}`),p(''));
      if(value.ctx.req===YEAR){body.push(coloredWordTable(page,p));continue;}
      for(const block of page.blocks){body.push(`<w:tbl><w:tblPr><w:tblW w:w="10640" w:type="dxa"/><w:tblBorders><w:top w:val="single" w:sz="4" w:color="D1B878"/><w:bottom w:val="single" w:sz="4" w:color="D1B878"/><w:left w:val="single" w:sz="4" w:color="D1B878"/><w:right w:val="single" w:sz="4" w:color="D1B878"/></w:tblBorders><w:tblCellMar><w:top w:w="100" w:type="dxa"/><w:left w:w="180" w:type="dxa"/><w:bottom w:w="100" w:type="dxa"/><w:right w:w="180" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid><w:gridCol w:w="10640"/></w:tblGrid><w:tr><w:trPr><w:cantSplit/></w:trPr><w:tc><w:tcPr><w:tcW w:w="10640" w:type="dxa"/></w:tcPr>${p(block.heading,true,10)}${block.lines.map(line=>p(line.text,line.bold)).join('')}</w:tc></w:tr></w:tbl>`,p(''));}
    }
    const ns='http://schemas.openxmlformats.org/wordprocessingml/2006/main';
    const documentXml=`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="${ns}" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:body>${body.join('')}<w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="600" w:right="800" w:bottom="600" w:left="800"/><w:footerReference w:type="default" r:id="footer"/></w:sectPr></w:body></w:document>`;
    return sgl5063ZipBlob([...media,{name:'[Content_Types].xml',data:'<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="jpg" ContentType="image/jpeg"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/footer.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/></Types>'},{name:'_rels/.rels',data:'<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="document" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>'},{name:'word/document.xml',data:documentXml},{name:'word/_rels/document.xml.rels',data:'<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="footer" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer.xml"/>'+logoRel+'</Relationships>'},{name:'word/footer.xml',data:`<?xml version="1.0"?><w:ftr xmlns:w="${ns}">${p('SGL 5.47 · Registro acumulativo')}</w:ftr>`}]);
  }
  async function summary(req,owner=target(),from='',to=''){
    if(!await beforeLeave())return;
    try{const ctx=context(req,owner),value=report(ctx,from,to),pages=reportPages(value);
      const range=req===YEAR?`<div class="actions"><label>Desde<input class="field" type="date" id="sgl531_from" value="${e(from)}"></label><label>Hasta<input class="field" type="date" id="sgl531_to" value="${e(to)}"></label>${btn('Filtrar',`SGL531.summary('${req}',${JSON.stringify(owner)},document.getElementById('sgl531_from').value,document.getElementById('sgl531_to').value)`)}</div>`:'';
      shell(ctx,value.title,`${range}<p>${value.keys.length} lecturas completadas. Los textos largos continúan en las páginas siguientes.</p><div class="sgl531-pages">${pages.map((page,i)=>`<article class="sgl531-sheet ${req===YEAR?'sgl535-colored-sheet':''}"><header><b>SISTEMA DE GESTIÓN DE MÚSICA</b><h3>${e(value.title)}</h3><p>${e(ctx.aspirant.fullName)} · ${e(req)}</p></header>${req===YEAR?`<div class="sgl535-card-grid" style="height:${page.used}pt">${page.blocks.map(block=>coloredCardHtml(block,page)).join('')}</div>`:page.blocks.map(block=>`<section><h4>${e(block.heading)}</h4>${block.lines.map(line=>`<div class="${line.bold?'label':''}">${e(line.text)||'&nbsp;'}</div>`).join('')}</section>`).join('')}<footer>${i+1} / ${pages.length}</footer></article>`).join('')}</div>`,btn('Volver al diario',`SGL531.open('${req}',${JSON.stringify(owner)})`)+btn('Visualizar PDF',`SGL531.preview('${req}',${JSON.stringify(owner)},${JSON.stringify(from)},${JSON.stringify(to)})`)+btn('Imprimir / Guardar PDF',`SGL531.preview('${req}',${JSON.stringify(owner)},${JSON.stringify(from)},${JSON.stringify(to)})`,true)+btn('Exportar PDF',`SGL531.exportFile('${req}',${JSON.stringify(owner)},'pdf',${JSON.stringify(from)},${JSON.stringify(to)})`,true)+btn('Exportar Word',`SGL531.exportFile('${req}',${JSON.stringify(owner)},'docx',${JSON.stringify(from)},${JSON.stringify(to)})`));
    }catch(error){toast(error.message);}
  }
  async function mobilePreview(value){
    const h=await reportHeader(value),pages=coloredPages(value);
    const scale=Math.min(1,(window.innerWidth-60)/816);
    return `<p class="muted">Vista de impresión. Usa Guardar PDF para descargar el documento.</p><div class="sgl536-preview">${pages.map((page,i)=>`<div style="height:${1056*scale}px;overflow:hidden;margin:12px 0"><article style="position:relative;box-sizing:border-box;background:white;color:black;width:612pt;height:792pt;transform:scale(${scale});transform-origin:top left;font-family:Arial"><header style="position:absolute;left:56pt;right:56pt;top:20pt;display:grid;grid-template-columns:48pt 1fr 48pt;gap:10pt;text-align:center"><img src="${e(h.official)}" style="width:48pt;height:43pt;object-fit:contain"><div><b style="font-size:8pt">${e(h.mission)}</b><div style="font-size:12pt;color:#1f739e;font-weight:bold">${e(h.title)}</div><span style="font-size:10pt">${e(h.name)}</span></div>${h.club?`<img src="${e(h.club)}" style="width:48pt;height:43pt;object-fit:contain">`:'<span></span>'}</header><div style="position:absolute;top:83pt;left:85pt;width:470pt;border-top:1.5pt solid #1f739e;padding-top:9pt"><b style="font-size:10pt">${e(value.title)}</b><div style="font-size:8pt;margin-top:3pt">${value.keys.length} lecturas · ${e(value.ctx.req)}</div></div><div class="sgl535-card-grid" style="position:absolute;left:85pt;top:132pt;width:470pt;height:600pt">${page.blocks.map(block=>coloredCardHtml(block,page)).join('')}</div><footer style="position:absolute;right:56pt;bottom:35pt;font-size:7pt">${i+1} / ${pages.length}</footer></article></div>`).join('')}</div>`;
  }
  async function preview(req,owner=target(),from='',to=''){
    if(!await beforeLeave())return;
    try{
      const ctx=context(req,owner),value=report(ctx,from,to),blob=await buildPdf(value);
      if(currentUser()?.id!==ctx.actor)return;
      if(previewUrl)URL.revokeObjectURL(previewUrl);
      previewUrl=URL.createObjectURL(blob);
      const call=JSON.stringify([req,owner,from,to]).slice(1,-1);
      const mobile=window.matchMedia('(max-width:820px)').matches;
      const mobilePages=mobile&&req===YEAR?await mobilePreview(value):'';
      shell(ctx,'Vista PDF · '+title(req,dataFor(ctx).choice),mobilePages||`<iframe id="sgl535_pdf" class="preview-frame" title="Vista PDF del informe" src="${previewUrl}#zoom=page-width" style="width:100%;height:65vh;border:0"></iframe>`,
        btn('Volver',`SGL531.summary(${call})`)+btn('Imprimir / Guardar PDF','SGL531.printPreview()',true)+btn('Guardar PDF',`SGL531.exportFile(${JSON.stringify(req)},${JSON.stringify(owner)},'pdf',${JSON.stringify(from)},${JSON.stringify(to)})`));
    }catch(error){toast(error.message);}
  }
  function printPreview(){
    const frame=document.getElementById('sgl535_pdf');
    if(!frame){if(previewUrl)window.open(previewUrl,'_blank','noopener');return;}
    try{frame.contentWindow.focus();frame.contentWindow.print();}catch{if(previewUrl)window.open(previewUrl,'_blank','noopener');}
  }
  async function exportFile(req,owner,ext='pdf',from='',to=''){
    if(!await beforeLeave())return;
    try{const ctx=context(req,owner),value=report(ctx,from,to),blob=await (ext==='pdf'?buildPdf(value):buildWord(value));if(currentUser()?.id!==ctx.actor)return;await saveBlobWithDialog(reportFileName({requirement:v3RequirementById(req,PROGRAM),aspirant:ctx.aspirant},ext),blob,ext==='pdf'?'Documento PDF|*.pdf':'Documento Word|*.docx');}catch(error){toast(error.message);}
  }

  root.SGL531={selectChapter,open,choose,clearChoice,certificate,start,day,pick,readPassage,closePassage,save,beforeLeave,send,summary,preview,printPreview,exportFile,context,dataFor,ready,report,reportPages,buildPdf,buildWord};
  const basePersist531=persist;
  persist=async function(...args){
    if(currentUser()?.role==='admin'){
      // Estos dos formatos pertenecen a la elección individual y al diario fijo.
      for(const key of [`${PROGRAM}:${YEAR}`,`${PROGRAM}:${FORTY}`]){
        for(const modes of [database.reportRequirementModes,database.settings?.reportRequirementModes]){
          if(modes&&Object.prototype.hasOwnProperty.call(modes,key)){
            database.settings||={};database.settings.musicDevotionPreviousModes531||={};
            database.settings.musicDevotionPreviousModes531[key]=clone(modes[key]);delete modes[key];
          }
        }
      }
    }
    return basePersist531.apply(this,args);
  };root.persist=persist;
  const baseOpen=openRequirement;
  openRequirement=function(req,...args){if(v3ActiveProgramId===PROGRAM&&[YEAR,FORTY].includes(req))return open(req);return baseOpen.call(this,req,...args);};root.openRequirement=openRequirement;
  const baseBible=renderBible;renderBible=function(...args){if(v3ActiveProgramId===PROGRAM){setContent('<div class="card"><h2>Devoción personal</h2><p>Elige Reavivados por su Palabra, Bible Plan o Subir certificado PDF para presentar tu lectura anual.</p><button class="btn primary" onclick="SGL531.open(\'LM-NUT-1\')">Abrir mi programa</button></div>');return open(YEAR);}return baseBible.apply(this,args);};root.renderBible=renderBible;
  const baseSend=sgl5049SendRequirementFromList;sgl5049SendRequirementFromList=function(req,...args){if(v3ActiveProgramId===PROGRAM&&[YEAR,FORTY].includes(req))return send(req);return baseSend.call(this,req,...args);};root.sgl5049SendRequirementFromList=sgl5049SendRequirementFromList;
  const basePresentTask=root.sgl5073PresentTask;
  root.sgl5073PresentTask=async function(id,...args){const task=database.tasks?.find(x=>x.id===id);if(task?.programId===PROGRAM&&[YEAR,FORTY].includes(task.requirementId)){try{const ctx=assertEditable(context(task.requirementId,task.assigneeId));if(!ready(ctx))return toast('Completa el diario o documento elegido antes de enviar esta tarea.');}catch(error){return toast(error.message);}}return basePresentTask.call(this,id,...args);};
  function exportContext(req){if([YEAR,FORTY].includes(req)&&v3SubmissionFor(String(modalOwnerId||target()),req,PROGRAM)?.primaryDocumentId)return null;const program=v3ModalProgramId||v3ActiveProgramId;if(program!==PROGRAM||![YEAR,FORTY].includes(req))return null;const ctx=context(req,String(modalOwnerId||target()));return dataFor(ctx).choice==='bible_plan'?null:ctx;}
  const basePdf=makeReportPdf;makeReportPdf=async function(req,...args){const ctx=exportContext(req);return ctx?buildPdf(report(ctx)):basePdf.call(this,req,...args);};root.makeReportPdf=makeReportPdf;
  const baseExportPdf=exportPdf;exportPdf=function(req,...args){const ctx=exportContext(req);return ctx?exportFile(req,ctx.owner,'pdf'):baseExportPdf.call(this,req,...args);};root.exportPdf=exportPdf;
  const baseExportWord=exportWord;exportWord=function(req,...args){const ctx=exportContext(req);return ctx?exportFile(req,ctx.owner,'docx'):baseExportWord.call(this,req,...args);};root.exportWord=exportWord;
  const basePreview=previewRequirementPdf;previewRequirementPdf=function(req,...args){const ctx=exportContext(req);return ctx?preview(req,ctx.owner):basePreview.call(this,req,...args);};root.previewRequirementPdf=previewRequirementPdf;
  const basePrint=printRequirement;printRequirement=async function(req,...args){const ctx=exportContext(req);if(!ctx)return basePrint.call(this,req,...args);if(!await beforeLeave())return;try{const blob=await buildPdf(report(ctx)),name=reportFileName({requirement:v3RequirementById(req,PROGRAM),aspirant:ctx.aspirant});const native=await nativePrintBlob(blob,name);if(native?.ok)return;if(previewUrl)URL.revokeObjectURL(previewUrl);previewUrl=URL.createObjectURL(blob);window.open(previewUrl,'_blank','noopener');}catch(error){toast(error.message);}};root.printRequirement=printRequirement;
  for(const name of ['closeModal','go','logout']){const base=root[name];root[name]=function(...args){if(!editor)return base.apply(this,args);return beforeLeave().then(ok=>ok?base.apply(this,args):false);};}
  document.addEventListener('visibilitychange',()=>{if(document.hidden)emergency();});
  window.addEventListener('beforeunload',event=>{if(dirty()){emergency();event.preventDefault();event.returnValue='';}});
  setInterval(()=>{if(editor?.editable&&!leavePending&&dirty())save(false);},10000);
})(globalThis);
