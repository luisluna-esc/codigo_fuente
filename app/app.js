'use strict';

let MASTER = window.__MASTER__ || null;
const APP_VERSION = '6.2.6-musica-online';
const OFFLINE_BUILD = '6.2.6-musica-online';
window.__SGL_OFFLINE_BUILD__ = OFFLINE_BUILD;
const IDB_NAME = 'ministerio_musica_online_v1';
const NATIVE_BRIDGE = 'http://127.0.0.1:49429';
const REMOTE_WEB = window.__MUSICA_WEB__===true || (location.protocol === 'https:' && !['localhost','127.0.0.1'].includes(location.hostname));
const REMOTE_ORIGIN = REMOTE_WEB ? location.origin+'/api/musica' : '';

const STATUS_LABELS = {
  pending:'Pendiente', assigned:'Asignado', in_progress:'En desarrollo',
  in_review:'Enviado a revisión', needs_changes:'Requiere corrección',
  approved:'Aprobado', rejected:'Rechazado'
};
function sgl548CanonicalRole(role){return role==='pastor_aspirant'?'aspirant':String(role||'').trim();}
function sgl548ViewAll(user=currentUser()){return ['admin','departamental','administrativo','secretaria','coordinador_musica','administrativo_sgm'].includes(user?.role);}
const ROLE_LABELS = {coordinador_sgm:'Coordinador SGM',administrativo_sgm:'Administrativo SGM',lider_musica:'Líder de Música',pastor_aspirant:'Pastor/Aspirante',administrativo:'Administrativo',admin:'Administrador', regional:'Regional', tutor:'Tutor', aspirant:'Aspirante'};
const REGIONAL_PROGRAM_IDS = ['music_ministry'];
const ACCOUNT_LABELS = {active:'Activo', inactive:'Inactivo / deshabilitado', archived:'Archivado'};
const TASK_LABELS = {pending:'Por presentar', presented:'Presentado', overdue:'Vencido'};
const MATERIAL_LABELS = {letters:'Cartas',reading_books:'Libros de lectura',support_books:'Libros de apoyo',videos:'Links de video'};
const DEFAULT_APP_LOGO = 'assets/ministerio_musica.png';
const DEFAULT_REPORT_LOGO = 'assets/simbolo-adventista.png';
const THEME_LABELS = {classic_blue:'Azul intenso',red:'Rojo intenso',sky:'Cielo institucional',violet:'Violeta sereno',coral:'Coral cálido',neutral:'Azul neutro'};
const AREA_SHORT = {'I':'I-G','II':'II-DE','III':'III-SA','IV':'IV-DA','V':'V-SAF','VI':'VI-OL','VII':'VII-EN','VIII':'VIII-AA','IX':'IX-EV'};
const CLASS_BRANDING = {
  A:{color:'#009ED3',color2:'#0076B5',soft:'#E3F7FD',logo:'assets/classes/embed/amigo.webp'},
  C:{color:'#E22918',color2:'#B5170D',soft:'#FDE9E6',logo:'assets/classes/embed/companero.webp'},
  E:{color:'#159A32',color2:'#087A22',soft:'#E6F7EA',logo:'assets/classes/embed/explorador.webp'},
  P:{color:'#4E5864',color2:'#242A31',soft:'#EDF0F3',logo:'assets/classes/embed/pionero.webp'},
  EX:{color:'#AC458E',color2:'#823369',soft:'#F7EAF3',logo:'assets/classes/embed/excursionista.webp'},
  G:{color:'#F29B16',color2:'#D97900',soft:'#FFF2D8',logo:'assets/classes/embed/guia.webp'}
};
function classBranding(item){return CLASS_BRANDING[item?.catalogKey||item?.id]||{color:'#355F8A',color2:'#203E61',soft:'#E9F0F7',logo:appLogo()};}
const CLASS_PAIRS = [
  {id:'A',regular:'Amigo',advanced:'Amigo de la Naturaleza',age:10},
  {id:'C',regular:'Compañero',advanced:'Compañero de Excursionismo',age:11},
  {id:'E',regular:'Explorador',advanced:'Explorador de Campo y Bosque',age:12},
  {id:'P',regular:'Pionero',advanced:'Pionero de Nuevas Fronteras',age:13},
  {id:'EX',regular:'Excursionista',advanced:'Excursionista en el Bosque',age:14},
  {id:'G',regular:'Guía',advanced:'Guía de Exploración',age:15}
];
const BIBLE_CANON = [
  ['GEN','Génesis',50],['EXO','Éxodo',40],['LEV','Levítico',27],['NUM','Números',36],['DEU','Deuteronomio',34],['JOS','Josué',24],['JDG','Jueces',21],['RUT','Rut',4],['1SA','1 Samuel',31],['2SA','2 Samuel',24],['1KI','1 Reyes',22],['2KI','2 Reyes',25],['1CH','1 Crónicas',29],['2CH','2 Crónicas',36],['EZR','Esdras',10],['NEH','Nehemías',13],['EST','Ester',10],['JOB','Job',42],['PSA','Salmos',150],['PRO','Proverbios',31],['ECC','Eclesiastés',12],['SNG','Cantares',8],['ISA','Isaías',66],['JER','Jeremías',52],['LAM','Lamentaciones',5],['EZK','Ezequiel',48],['DAN','Daniel',12],['HOS','Oseas',14],['JOL','Joel',3],['AMO','Amós',9],['OBA','Abdías',1],['JON','Jonás',4],['MIC','Miqueas',7],['NAM','Nahúm',3],['HAB','Habacuc',3],['ZEP','Sofonías',3],['HAG','Hageo',2],['ZEC','Zacarías',14],['MAL','Malaquías',4],
  ['MAT','Mateo',28],['MRK','Marcos',16],['LUK','Lucas',24],['JHN','Juan',21],['ACT','Hechos',28],['ROM','Romanos',16],['1CO','1 Corintios',16],['2CO','2 Corintios',13],['GAL','Gálatas',6],['EPH','Efesios',6],['PHP','Filipenses',4],['COL','Colosenses',4],['1TH','1 Tesalonicenses',5],['2TH','2 Tesalonicenses',3],['1TI','1 Timoteo',6],['2TI','2 Timoteo',4],['TIT','Tito',3],['PHM','Filemón',1],['HEB','Hebreos',13],['JAS','Santiago',5],['1PE','1 Pedro',5],['2PE','2 Pedro',3],['1JN','1 Juan',5],['2JN','2 Juan',1],['3JN','3 Juan',1],['JUD','Judas',1],['REV','Apocalipsis',22]
];
let flatRequirements = MASTER ? MASTER.areas.flatMap(area => area.requirements.map(requirement => ({...requirement, areaCode:area.code, areaTitle:area.title}))) : [];

let database;
let idb;
let currentPage = 'dashboard';
let activeAspirantId = null;
let modalOwnerId = null;
let modalRequirementId = null;
let syncTimer = null;
let draftTimer = null;
let remoteHeartbeatTimer = null;
let snapshotTimer = null;
let snapshotWriting = false;
let snapshotAgain = false;
let mandatoryLogoDataUrl = '';
let taskReminderShown = false;
let requirementStatusFilter = '';
let activityRange = 'today';
let leadershipRange = {from:'',to:''};
let profileStudioImage = null;
let profileStudioRotation = 0;
let profileStudioOffset = {x:0,y:0};
let profileStudioDrag = null;
let photoStudioTarget = null;
let materialPreviewUrl = '';
let classCatalog = {};
let activeClassId = '';
let activeClassSection = 'dashboard';
let classMemberSearch = '';
let classMemberFilter = 'active';
let classTrackFilter = 'all';
let classPlanFilter = 'all';
let sglHistoryReady = false;
let sglHistoryRestoring = false;
let sglImageViewer = null;
let classAttendanceDate = '';
const reportPdfCache = new Map();

function newId(prefix='id') {
  return `${prefix}_${crypto.randomUUID ? crypto.randomUUID() : `${Date.now()}_${Math.random().toString(16).slice(2)}`}`;
}

function isoNow() { return new Date().toISOString(); }
function localDate() { const now=new Date(), offset=now.getTimezoneOffset(); return new Date(now.getTime()-offset*60000).toISOString().slice(0,10); }
function esc(value) { return String(value ?? '').replace(/[&<>"']/g, char => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char])); }
function short(value, length=110) { value = String(value ?? ''); return value.length > length ? `${value.slice(0, length - 1)}…` : value; }
function initials(value) { return String(value || '?').trim().split(/\s+/).slice(0, 2).map(part => part[0]?.toUpperCase() || '').join(''); }
function normalize(value) { return String(value || '').trim().toLowerCase(); }
function nameAutocompleteHistory(){return Array.isArray(database?.settings?.nameAutocompleteHistory)?database.settings.nameAutocompleteHistory:[];}
function nameAutocompleteOptions(){return nameAutocompleteHistory().map(name=>`<option value="${esc(name)}"></option>`).join('');}
function rememberAutocompleteName(name){const clean=String(name||'').replace(/\s+/g,' ').trim();if(!clean)return;database.settings.nameAutocompleteHistory=[clean,...nameAutocompleteHistory().filter(item=>normalize(item)!==normalize(clean))].slice(0,60);}
async function clearNameAutocompleteHistory(){database.settings.nameAutocompleteHistory=[];await persist(true,{action:'Autocompletado de nombres limpiado',entity:'settings',detail:'Historial de nombres'});for(const root of document.querySelectorAll('#class_name_history,#user_name_history'))root.innerHTML='';toast('Nombres guardados para autocompletar eliminados.');}
function currentUser() { let id='';try{id=sessionStorage.getItem('musica_session_user')||(!REMOTE_WEB?localStorage.getItem('musica_session_user'):'')||'';}catch{}return database?.users?.find(user => user.id === id && (!window.SGL593Allows || window.SGL593Allows(database,user,true))) || null; }
function clubFor(user) { return database.clubs.find(club => club.id === user?.clubId) || null; }
function userById(id) { return database.users.find(user => user.id === id) || null; }
function appLogo() { return database?.settings?.musicAppLogo || DEFAULT_APP_LOGO; }
function sglShouldShowOfflineMode() {
  const message=String(database?.sync?.message||'').toLowerCase();
  // SGL 5.0.77: en acceso HTTPS remoto (especialmente *.ts.net),
  // navigator.onLine es solo una pista del navegador. El estado Offline se
  // muestra cuando el servidor fue realmente confirmado como inaccesible.
  if(REMOTE_WEB){
    if(window.__SGL_REMOTE_VERIFIED_OFFLINE__===true)return true;
    if(window.__SGL_REMOTE_VERIFIED_OFFLINE__===false)return false;
    return /sin conexi[oó]n|offline|servidor no disponible/.test(message);
  }
  if (!navigator.onLine) return true;
  return /sin conexi[oó]n|offline|servidor no disponible|conexi[oó]n m[oó]vil inestable/.test(message);
}
function sglApplyOfflineMode(force=false,message='') {
  const active=!!force || sglShouldShowOfflineMode();
  if(document.body)document.body.classList.toggle('musica-offline-mode',active);
  const text=message || 'SGL está trabajando en modo offline. Tus cambios se guardan en este dispositivo y se sincronizarán al volver la conexión.';
  const existing=document.getElementById('sgl_offline_banner');
  if(!active){ existing?.remove(); return; }
  if(existing){ existing.querySelector('span').textContent=text; return; }
  const target=document.querySelector('.auth-card,.content');
  if(target)target.insertAdjacentHTML('afterbegin',`<div id="sgl_offline_banner" class="notice offline-mode"><b>Modo offline</b><span>${esc(text)}</span><button type="button" onclick="sglRefreshWithoutNavigation()">Actualizar vista</button></div>`);
}
let sglRemoteConnectivityProbe=null;
function sglNotifyServiceWorkerConnectivity(type,extra={}){
  if(!('serviceWorker' in navigator))return;
  const message={type,...extra};
  try{navigator.serviceWorker.controller?.postMessage(message);}catch{}
  try{navigator.serviceWorker.getRegistration?.().then(reg=>{try{(reg?.active||reg?.waiting)?.postMessage(message);}catch{}}).catch(()=>{});}catch{}
}
function sglClearOfflineModeVerified(message='Conectado al servidor'){
  window.__SGL_REMOTE_VERIFIED_OFFLINE__=false;
  sglNotifyServiceWorkerConnectivity('SERVER_REACHABLE');
  const current=String(database?.sync?.message||'');
  if(database?.sync&&/sin conexi[oó]n|offline|servidor no disponible|conexi[oó]n m[oó]vil inestable|verificando servidor/i.test(current)){
    const pending=Array.isArray(database.syncQueue)&&database.syncQueue.some(item=>['queued','retry','sending'].includes(item?.status));
    database.sync={...(database.sync||{}),state:pending?'pending':'ok',message:pending?'Conexión recuperada · sincronizando cambios':message,lastSyncedAt:database.sync?.lastSyncedAt||null};
  }
  document.body?.classList.remove('musica-offline-mode');
  document.getElementById('sgl_offline_banner')?.remove();
  try{renderTopSync?.();}catch{}
}
async function sglProbeRemoteServerOnce(timeoutMs=6000){
  if(!REMOTE_WEB)return navigator.onLine;
  const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),Math.max(1000,Number(timeoutMs)||6000));
  try{
    const response=await fetch(`${REMOTE_ORIGIN}/api/remote/info`,{method:'GET',cache:'no-store',headers:{'Cache-Control':'no-cache','Pragma':'no-cache'},signal:controller.signal});
    return !!response?.ok;
  }catch{return false;}finally{clearTimeout(timer);}
}
async function sglConfirmRemoteServer(options={}){
  if(!REMOTE_WEB)return navigator.onLine;
  if(sglRemoteConnectivityProbe)return sglRemoteConnectivityProbe;
  const attempts=Math.max(1,Number(options.attempts)||2),timeouts=Array.isArray(options.timeouts)&&options.timeouts.length?options.timeouts:[4500,7500],delay=Math.max(100,Number(options.delay)||600);
  sglRemoteConnectivityProbe=(async()=>{
    for(let index=0;index<attempts;index++){
      if(await sglProbeRemoteServerOnce(timeouts[Math.min(index,timeouts.length-1)])){sglClearOfflineModeVerified();return true;}
      if(index<attempts-1)await new Promise(resolve=>setTimeout(resolve,delay));
    }
    window.__SGL_REMOTE_VERIFIED_OFFLINE__=true;
    sglNotifyServiceWorkerConnectivity('SERVER_UNREACHABLE',{forMs:6000});
    return false;
  })().finally(()=>{sglRemoteConnectivityProbe=null;});
  return sglRemoteConnectivityProbe;
}
window.sglConfirmRemoteServer=sglConfirmRemoteServer;
function authConnectionNotice(message='') {
  const offline=!navigator.onLine;
  const parts=[];
  if(offline)parts.push(`<div class="notice offline-mode"><b>Modo offline:</b><span> No hay conexión. Puedes usar la copia ya guardada; para iniciar por primera vez necesitas conectar con el servidor.</span></div>`);
  if(message)parts.push(`<div class="notice warn">${esc(message)}</div>`);
  return parts.join('');
}
function requirementAreaCode(requirement={}) {
  return String(requirement.areaCode || (String(requirement.areaTitle || requirement.area || '').match(/^([IVXLCDM]+)/i)||[])[1] || '').toUpperCase();
}
function requirementAreaAbbreviation(requirement={}) {
  const code=requirementAreaCode(requirement);
  return AREA_SHORT[code] || requirement.areaShort || code || String(requirement.id || '').split('-')[0] || 'REQ';
}
function requirementAreaBadge(requirement={}) {
  return `<div class="req-code" title="${esc(requirement.areaTitle || requirement.area || requirement.id || '')}">${esc(requirementAreaAbbreviation(requirement))}</div>`;
}

function openIndexedDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(IDB_NAME, 1);
    request.onupgradeneeded = () => {
      const store = request.result;
      if (!store.objectStoreNames.contains('kv')) store.createObjectStore('kv');
      if (!store.objectStoreNames.contains('files', {keyPath:'id'})) store.createObjectStore('files', {keyPath:'id'});
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

function txRequest(store, mode, operation) {
  return new Promise((resolve, reject) => {
    const transaction = idb.transaction(store, mode);
    const request = operation(transaction.objectStore(store));
    let result;
    request.onsuccess = () => { result=request.result; };
    request.onerror = () => reject(request.error);
    transaction.oncomplete = () => resolve(result);
    transaction.onabort = () => reject(transaction.error || new Error('No se confirmó la escritura local.'));
    transaction.onerror = () => reject(transaction.error || request.error);
  });
}

const kvGet = key => txRequest('kv', 'readonly', store => store.get(key));
const kvSet = (key, value) => txRequest('kv', 'readwrite', store => store.put(value, key));
const filePut = value => txRequest('files', 'readwrite', store => store.put(value));
const fileGet = id => txRequest('files', 'readonly', store => store.get(id));
const fileDelete = id => txRequest('files', 'readwrite', store => store.delete(id));
const allFiles = () => txRequest('files', 'readonly', store => store.getAll());
const clearFiles = () => txRequest('files', 'readwrite', store => store.clear());

async function sglWorkerInfo(worker,timeoutMs=700) {
  if(!worker)return null;
  return new Promise(resolve=>{
    const channel=new MessageChannel();
    let done=false;
    const finish=value=>{if(done)return;done=true;clearTimeout(timer);resolve(value||null);};
    const timer=setTimeout(()=>finish(null),timeoutMs);
    channel.port1.onmessage=event=>finish(event.data||null);
    try{worker.postMessage({type:'SGL_VERSION'},[channel.port2]);}catch{finish(null);}
  });
}
async function sglIsCurrentServiceWorkerController() {
  const info=await sglWorkerInfo(navigator.serviceWorker?.controller,650);
  return info?.currentVersion===APP_VERSION||info?.release===APP_VERSION||info?.version===APP_VERSION;
}
window.sglIsCurrentServiceWorkerController=sglIsCurrentServiceWorkerController;

async function sglWaitForCurrentServiceWorker(registration,timeoutMs=90000) {
  if(!registration)return null;
  const deadline=Date.now()+timeoutMs;
  try{await registration.update();}catch{}
  while(Date.now()<deadline){
    for(const worker of [registration.active,registration.waiting]){
      const info=await sglWorkerInfo(worker,650);
      if(info?.currentVersion===OFFLINE_BUILD||info?.release===OFFLINE_BUILD||info?.version===OFFLINE_BUILD){
        if(worker===registration.waiting)try{worker.postMessage({type:'SKIP_WAITING',version:OFFLINE_BUILD});}catch{}
        return worker;
      }
    }
    const installing=registration.installing;
    if(installing?.state==='installed')try{installing.postMessage({type:'SKIP_WAITING',version:OFFLINE_BUILD});}catch{}
    await new Promise(resolve=>setTimeout(resolve,180));
  }
  return null;
}

async function sglWaitForServiceWorkerControl(timeoutMs=5000) {
  if(!('serviceWorker' in navigator))return false;
  const deadline=Date.now()+timeoutMs;
  while(Date.now()<deadline){
    if(await sglIsCurrentServiceWorkerController())return true;
    await new Promise(resolve=>setTimeout(resolve,120));
  }
  return sglIsCurrentServiceWorkerController();
}
async function sglRequestPersistentStorage() {
  const record={supported:false,persisted:false,requested:false,checkedAt:isoNow()};
  try {
    if (!navigator.storage?.persisted || !navigator.storage?.persist) {
      record.reason='StorageManager persist() no disponible';
      localStorage.setItem('sgl_storage_persistence',JSON.stringify(record));
      return record;
    }
    record.supported=true;
    record.persisted=await navigator.storage.persisted();
    if(!record.persisted){record.requested=true;record.persisted=await navigator.storage.persist();}
    if(navigator.storage.estimate){const estimate=await navigator.storage.estimate();record.usage=Number(estimate?.usage||0);record.quota=Number(estimate?.quota||0);}
    if(!record.persisted)record.reason='El navegador no concedio almacenamiento persistente';
    localStorage.setItem('sgl_storage_persistence',JSON.stringify(record));
  } catch (error) {
    record.reason=String(error?.message||error||'Error de almacenamiento');
    try{localStorage.setItem('sgl_storage_persistence',JSON.stringify(record));}catch{}
  }
  return record;
}
window.sglRequestPersistentStorage=sglRequestPersistentStorage;

async function sglVerifyOfflineCache(worker=navigator.serviceWorker?.controller,timeoutMs=5000) {
  if(!worker)return {ok:false,complete:false,error:'No Service Worker controller'};
  return new Promise(resolve=>{
    const channel=new MessageChannel();
    const timer=setTimeout(()=>resolve({ok:false,complete:false,timeout:true}),timeoutMs);
    channel.port1.onmessage=event=>{clearTimeout(timer);resolve(event.data||{ok:false,complete:false});};
    try{worker.postMessage({type:'VERIFY_OFFLINE',version:OFFLINE_BUILD},[channel.port2]);}
    catch(error){clearTimeout(timer);resolve({ok:false,complete:false,error:String(error?.message||error)});}
  });
}
window.sglVerifyOfflineCache=sglVerifyOfflineCache;

let sglOfflinePreparationPromise=null;
async function sglPrepareOfflineShellOnce() {
  const supported = 'serviceWorker' in navigator && (window.__MUSICA_WEB__ || location.protocol === 'https:') && !['localhost','127.0.0.1'].includes(location.hostname);
  if (!supported) return false;
  try {
    const registration = await (window.__SGL_SW_REGISTRATION__ || navigator.serviceWorker.ready);
    const ready = registration || await navigator.serviceWorker.ready;
    // Nunca enviar WARM_OFFLINE a un worker de otra release: la cache de arranque
    // 5.0.43 es atomica y solo se valida contra el Service Worker actual.
    const worker = await sglWaitForCurrentServiceWorker(ready,90000);
    if (!worker) return false;
    const result = await new Promise(resolve => {
      const channel = new MessageChannel();
      const timer = setTimeout(() => resolve({ok:false,timeout:true}),60000);
      channel.port1.onmessage = event => { clearTimeout(timer); resolve(event.data || {ok:false}); };
      worker.postMessage({type:'WARM_OFFLINE',version:OFFLINE_BUILD},[channel.port2]);
    });
    if (!result?.ok || !result?.complete || (result?.missing||[]).length) {
      // No se borra un marcador previo por un fallo transitorio. El estado real
      // se revalida contra Cache Storage + Service Worker al mostrar/arrancar.
      try{window.sgl509UpdateOfflineIndicator?.();}catch{}
      return false;
    }

    const controlled=await sglWaitForServiceWorkerControl(30000);
    if(!controlled){
      try{window.sgl509UpdateOfflineIndicator?.();}catch{}
      return false;
    }

    const verification=await sglVerifyOfflineCache(navigator.serviceWorker.controller,5000);
    if(!verification?.ok || !verification?.complete){
      try{window.sgl509UpdateOfflineIndicator?.();}catch{}
      return false;
    }

    // Android/Chrome puede negar persist() aunque Service Worker + Cache Storage
    // funcionen correctamente. Se solicita y registra como proteccion adicional,
    // pero no se usa para bloquear el arranque offline.
    const persistence=await sglRequestPersistentStorage();
    const persistent=!!persistence?.persisted;
    try {
      localStorage.setItem('sgl_offline_ready',JSON.stringify({
        version:APP_VERSION,
        build:OFFLINE_BUILD,
        at:isoNow(),
        controlled:true,
        cacheComplete:true,
        storagePersisted:persistent,
        storagePersistenceSupported:!!persistence?.supported,
        required:Number(verification.required||result.required||0)
      }));
      sessionStorage.removeItem('sgl_offline_control_reload');
    } catch {}
    try{window.sgl509UpdateOfflineIndicator?.();}catch{}
    return true;
  } catch {
    // Un corte de red durante una revalidacion no des-prepara una copia que ya
    // estaba completa; la siguiente comprobacion usa evidencia real del cache.
    return false;
  }
}
async function sglPrepareOfflineShell() {
  if(sglOfflinePreparationPromise)return sglOfflinePreparationPromise;
  sglOfflinePreparationPromise=sglPrepareOfflineShellOnce().finally(()=>{sglOfflinePreparationPromise=null;});
  return sglOfflinePreparationPromise;
}
window.sglPrepareOfflineShell = sglPrepareOfflineShell;

function blankDatabase() {
  return {
    schemaVersion:6,
    version:APP_VERSION,
    revision:0,
    serverVersion:0,
    installationId:newId('install'),
    createdAt:isoNow(),
    updatedAt:isoNow(),
    settings:{serverUrl:'', serverToken:'', autoSync:true, backupFrequency:'weekly',reportFont:'Arial',theme:'classic_blue',nameAutocompleteHistory:[]},
    clubs:[], users:[], classes:defaultClasses(), progress:{}, configs:{}, indicationRules:[], tasks:[], materials:[], bibleVersions:[], audit:[], errors:[], conflicts:[],
    sync:{state:'local', lastSyncedAt:null, message:'Guardado en este dispositivo'}
  };
}

function defaultClasses(){return [];}

function applyTheme(){const user=currentUser?.();let theme=user?.theme||'classic_blue';if(!THEME_LABELS[theme])theme=THEME_LABELS[user?.lastLightTheme]?user.lastLightTheme:'classic_blue';if(user){user.theme=theme;if('lastLightTheme' in user)delete user.lastLightTheme;}document.documentElement.dataset.theme=theme;}

async function inflateEmbeddedJson(source){const bytes=Uint8Array.from(atob(source||''),char=>char.charCodeAt(0)),stream=new Blob([bytes]).stream().pipeThrough(new DecompressionStream('gzip'));return JSON.parse(await new Response(stream).text());}
async function loadEmbeddedCatalogs(){
  if(!MASTER)MASTER=await inflateEmbeddedJson(window.__MASTER_GZIP__);
  flatRequirements=MASTER.areas.flatMap(area=>area.requirements.map(requirement=>({...requirement,areaCode:area.code,areaTitle:area.title})));
  try{
    if(window.__CLASS_DATA_GZIP__) classCatalog=await inflateEmbeddedJson(window.__CLASS_DATA_GZIP__);
    else if(window.__CLASS_CATALOG__) classCatalog=window.__CLASS_CATALOG__;
    else {
      const response=await fetch('class_catalog.json',{cache:'no-store'});
      if(!response.ok)throw new Error(`Catálogo de clases HTTP ${response.status}`);
      classCatalog=await response.json();
    }
  }
  catch(error){classCatalog={};await logClientError('Catálogo de clases',error);}
}

async function init() {
  idb = await openIndexedDB();
  await loadEmbeddedCatalogs();
  const browserDatabase=await kvGet('database'),diskSnapshot=await readNativeSnapshot(),diskDatabase=diskSnapshot?.database;
  const browserGeneration=Math.max(1,Number(browserDatabase?.restoreGeneration||1)||1),diskGeneration=Math.max(1,Number(diskDatabase?.restoreGeneration||1)||1),preferDisk=diskDatabase&&(!browserDatabase||diskGeneration>browserGeneration||(diskGeneration===browserGeneration&&((diskDatabase.users?.length&&!browserDatabase.users?.length)||Number(diskDatabase.revision||0)>Number(browserDatabase.revision||0)))); // restoreGeneration before revision
  database = preferDisk ? diskDatabase : browserDatabase || diskDatabase || blankDatabase();
  if(preferDisk&&Array.isArray(diskSnapshot.files)){await clearFiles();for(const file of diskSnapshot.files)await filePut(file);}
  migrateDatabase();
  await normalizeSavedLogos();
  installSafetyHandlers();
  await persist(false);
  window.addEventListener('offline',()=>window.MusicaOnline?.lockConnection());
  window.addEventListener('online',()=>window.MusicaOnline?.checkConnection());
  routeInitial();
}


// SGL 5.22 - migración conservadora de indicaciones predeterminadas de Clase Agrupada.
// Solo reemplaza reglas globales que todavía contienen el valor predeterminado anterior;
// una regla personalizada por el Administrador, Club, rol o persona se conserva intacta.
const SGL522_CLASS_GROUPED_PREVIOUS_INDICATIONS=Object.freeze({"I-01":"Certificado de miembro activo del club","I-02":"Memorización y explicación","I-03":"Informe creativo","I-04":"Comprensión de una de las opciones","I-05":"Memorización y explicación","I-06":"Memorización y explicación","I-07":"Memorización y explicación","I-08":"Demostrar conocimiento","I-09":"Demostrar conocimiento","I-10":"Demostrar conocimiento","I-11":"Demostrar conocimiento y resumen","I-12":"Demostrar conocimiento y resumen","I-13":"Demostrar conocimiento y resumen","I-14":"Demostrar conocimiento y resumen","I-15":"Demostrar conocimiento y resumen","I-16":"Demostrar conocimiento y resumen","I-17":"Certificado otorgado por el Club","II-01":"Memorización y conocimiento del tema.","II-02":"Memorización y conocimiento del tema.","II-03":"Memorización y conocimiento del tema.","II-04":"Memorización y conocimiento del tema.","II-05":"Memorización y conocimiento del tema.","II-06":"Memorización y conocimiento del tema.","II-07":"Explicar correctamente","II-08":"Explicar correctamente","II-09":"Explicar correctamente","II-10":"Explicar correctamente","II-11":"Explicar correctamente","II-12":"Explicar correctamente","II-13":"Leer cada pasaje, escoger el versículo que más le gustó y escribir de uno a tres párrafos sobre lo que más le llamó la atención.","II-14":"Leer cada pasaje, escoger el versículo que más le gustó y escribir de uno a tres párrafos sobre lo que más le llamó la atención.","II-15":"Leer cada pasaje, escoger el versículo que más le gustó y escribir de uno a tres párrafos sobre lo que más le llamó la atención.","II-16":"Leer cada pasaje, escoger el versículo que más le gustó y escribir de uno a tres párrafos sobre lo que más le llamó la atención.","II-17":"Leer cada pasaje, escoger el versículo que más le gustó y escribir de uno a tres párrafos sobre lo que más le llamó la atención.","II-18":"Leer cada pasaje, escoger el versículo que más le gustó y escribir de uno a tres párrafos sobre lo que más le llamó la atención.","II-19":"Participación y conocimiento del tema","II-20":"Explicar correctamente","II-21":"Participación y conocimiento del tema","II-22":"Participación y conocimiento del tema","II-22.1":"Explicación correcta de acuerdo a la opción escogida.","II-23":"Participación y conocimiento del tema","II-23.1":"Explicación correcta de acuerdo a la opción escogida.","II-24":"Invitar a un pastor y el líder debe entender el tema","II-25":"Invitar a un pastor y el líder debe entender el tema","II-26":"Invitar a un pastor y el líder debe entender el tema","II-27":"Registro fotográfico y datos del invitado","II-28":"Invitar a un pastor y el líder debe entender el tema","II-29":"Exposición"});
function migrateClassGroupedIndications522(){
  database.migrations ||= {};
  if(database.migrations.classGroupedIndications522)return;
  const normalize=value=>String(value??'').replace(/\s+/g,' ').trim();
  const defaults=new Map((flatRequirements||[]).map(item=>[String(item.id||''),String(item.tutor_instructions_default||'')]));
  let updatedRules=0,preservedCustomRules=0;
  for(const rule of (database.indicationRules||[])){
    const programId=String(rule.programId||'class_grouped'),requirementId=String(rule.requirementId||'');
    if(programId!=='class_grouped'||String(rule.scope||'all')!=='all'||!(requirementId in SGL522_CLASS_GROUPED_PREVIOUS_INDICATIONS))continue;
    const current=normalize(rule.config?.instructions),previous=normalize(SGL522_CLASS_GROUPED_PREVIOUS_INDICATIONS[requirementId]);
    if(current&&current!==previous){preservedCustomRules++;continue;}
    const next=defaults.get(requirementId);if(!next)continue;
    rule.config={...(rule.config||{}),instructions:next};rule.active=rule.active!==false;rule.updatedAt=isoNow();updatedRules++;
  }
  database.migrations.classGroupedIndications522={at:isoNow(),defaultsUpdated:Object.keys(SGL522_CLASS_GROUPED_PREVIOUS_INDICATIONS).length,updatedRules,preservedCustomRules};
}

function migrateDatabase() {
  database.users ||= [];
  database.clubs ||= [];
  database.progress ||= {};
  database.configs ||= {};
  database.indicationRules ||= [];
  database.tasks ||= [];
  database.materials ||= [];
  database.classes ||= defaultClasses();
  database.bibleVersions ||= [];
  database.audit ||= [];
  database.errors ||= [];
  database.conflicts ||= [];
  database.settings ||= {serverUrl:'', serverToken:'', autoSync:true, backupFrequency:'weekly',reportFont:'Arial'};
  database.settings.reportFont ||= 'Arial';
  database.settings.theme ||= 'classic_blue';
  database.settings.nameAutocompleteHistory = Array.isArray(database.settings.nameAutocompleteHistory) ? database.settings.nameAutocompleteHistory : [];
  database.sync ||= {state:'local', lastSyncedAt:null, message:'Guardado en este dispositivo'};
  for (const user of database.users) {
    if (user.role === 'leader') user.role = 'aspirant';
    user.accountState ||= user.active === false ? 'inactive' : 'active';
    user.active = user.accountState === 'active';
    if(user.role==='tutor'&&'canCreateClasses' in user)delete user.canCreateClasses;
    user.theme ||= 'classic_blue';
  }
  for (const user of database.users) user.reportFont ||= database.settings.reportFont || 'Arial';
  const previousClasses=Array.isArray(database.classes)?database.classes:[],official=CLASS_PAIRS.map((definition,index)=>{const previous=previousClasses.find(item=>item.id===definition.id)||{};return {...definition,catalogKey:definition.id,name:`${definition.regular} + ${definition.advanced}`,active:previous.active!==undefined?previous.active:index===0,assignedUserIds:Array.isArray(previous.assignedUserIds)?previous.assignedUserIds:[],work:previous.work||{members:[],planning:{},attendance:[],tracking:{}},createdAt:previous.createdAt||isoNow(),updatedAt:previous.updatedAt||null};}),custom=previousClasses.filter(item=>item.custom&&!CLASS_PAIRS.some(definition=>definition.id===item.id));
  database.classes=[...official,...custom];
  database.schemaVersion = 6;
  database.version = APP_VERSION;
  migrateClassGroupedIndications522();
  for (const user of database.users.filter(user => ['aspirant','pastor_aspirant'].includes(user.role))) ensureProgress(user.id);
  for(const item of database.classes){item.assignedUserIds ||= [];item.active=item.active!==false;classWork(item);}
  applyTheme();
}

function installSafetyHandlers() {
  const record=(kind,error)=>logClientError(kind,error?.reason||error?.error||error);
  window.addEventListener('error',event=>record('Error',event));
  window.addEventListener('unhandledrejection',event=>record('Promesa rechazada',event));
  window.addEventListener('beforeunload',saveEmergencyDraft);
  document.addEventListener('visibilitychange',()=>{if(document.hidden)saveEmergencyDraft();else if(typeof window.v3WakeSync==='function')window.v3WakeSync('visible');});
  window.addEventListener('pageshow',()=>{if(typeof window.v3WakeSync==='function')window.v3WakeSync('pageshow');});
  window.addEventListener('focus',()=>{if(typeof window.v3WakeSync==='function')window.v3WakeSync('focus');});
  document.addEventListener('keydown',event=>{if(event.key==='Escape'&&document.querySelector('.modal-bg'))closeModal();});
}

async function logClientError(kind,error) {
  try {
    const message=String(error?.message||error||'Error desconocido'), stack=String(error?.stack||'').slice(0,4000);
    database.errors.unshift({id:newId('err'),at:isoNow(),userId:currentUser()?.id||null,kind,message,stack});
    database.errors=database.errors.slice(0,100);
    await kvSet('database',database);
  } catch {}
}

async function persist(markPending=true, auditEvent=null) {
  database.updatedAt = isoNow();
  database.revision = (database.revision || 0) + 1;
  if (auditEvent) addAudit(auditEvent.action, auditEvent.entity, auditEvent.detail);
  if (markPending) {
    const pendingProposal=database.sync?.pendingProposal||null;
    database.sync = {
      state:database.settings.serverUrl ? 'pending' : 'local',
      lastSyncedAt:database.sync?.lastSyncedAt || null,
      message:database.settings.serverUrl ? 'Cambio guardado · pendiente de sincronización' : 'Guardado en este dispositivo',
      pendingProposal,
      needsUpload:true
    };
  }
  await kvSet('database', database);
  scheduleNativeSnapshot();
  renderTopSync();
  if (markPending && database.settings.autoSync && (REMOTE_WEB || navigator.onLine) && database.settings.serverUrl) {
    clearTimeout(syncTimer);
    syncTimer = setTimeout(syncNow, 1400);
  }
}

// Guarda borradores/evidencias únicamente en este dispositivo. No crea auditoría,
// no cambia el estado de sincronización y nunca genera una solicitud de aprobación.
async function persistDraftLocal() {
  database.updatedAt = isoNow();
  database.sync = {...(database.sync||{}),hasLocalDraft:true};
  await kvSet('database', database);
  scheduleNativeSnapshot();
}

function addAudit(action, entity='', detail='') {
  const actor = currentUser();
  database.audit.unshift({id:newId('audit'), at:isoNow(), actorId:actor?.id || null, actorName:actor?.fullName || 'Sistema', action, entity, detail});
  database.audit = database.audit.slice(0, 1000);
}

function routeInitial() {
  if (REMOTE_WEB) {
    const remoteUsers = database.users.filter(user => ['admin','departamental','administrativo','secretaria','coordinador_musica','regional','director','director_asociado','tutor','instructor','aspirant','pastor_aspirant','administrativo_sgm','lider_musica'].includes(user.role) && (user.accountState||'active') === 'active');
    if (!remoteUsers.length) return renderRemoteLogin();
    if (!currentUser()) return renderLogin();
    return renderShell();
  }
  const hasAdmin = database.users.some(user => user.role === 'admin' && user.active !== false);
  if (!hasAdmin) return renderBootstrap();
  if (!currentUser()) return renderLogin();
  renderShell();
}

const SGL_LOGIN_MESSAGES = [
  {kind:'Biblia',text:'El liderazgo cristiano comienza sirviendo, no buscando una posición.',ref:'Marcos 10:43-45'},
  {kind:'Biblia',text:'Quien guía a otros también debe ser ejemplo en palabra, conducta, amor, fe y pureza.',ref:'1 Timoteo 4:12'},
  {kind:'Biblia',text:'La sabiduría para dirigir se pide a Dios con confianza.',ref:'Santiago 1:5'},
  {kind:'Biblia',text:'El líder fuerte no camina solo: cobra ánimo porque Dios va con él.',ref:'Josué 1:6-9'},
  {kind:'Biblia',text:'El trabajo hecho para Dios merece entrega de corazón.',ref:'Colosenses 3:23'},
  {kind:'Biblia',text:'Antes de dirigir, aprende a cuidar bien aquello que Dios puso bajo tu responsabilidad.',ref:'1 Timoteo 3:4-5'},
  {kind:'Biblia',text:'El buen pastor conoce, cuida y protege a quienes le fueron confiados.',ref:'Juan 10:11-14'},
  {kind:'Biblia',text:'Los planes se fortalecen cuando se escuchan buenos consejos.',ref:'Proverbios 15:22'},
  {kind:'Biblia',text:'La humildad abre el camino para servir mejor a los demás.',ref:'Filipenses 2:3-5'},
  {kind:'Biblia',text:'La perseverancia en la oración sostiene las responsabilidades del servicio.',ref:'Romanos 12:11-12'},
  {kind:'Sábado',text:'El sábado recuerda al líder que su valor no depende de producir más, sino de descansar y confiar en Dios.',ref:'Éxodo 20:8-11'},
  {kind:'Sábado',text:'El sábado es también una oportunidad para hacer el bien y aliviar cargas.',ref:'Mateo 12:11-12'},
  {kind:'Sábado',text:'Haz del sábado un día de comunión, servicio y alegría en el Señor.',ref:'Isaías 58:13-14'},
  {kind:'Elena G. de White',text:'Ante cada dificultad, quien lleva responsabilidades puede verla como una invitación a buscar a Dios en oración.',ref:'Christian Leadership, cap. Christian Leadership'},
  {kind:'Elena G. de White',text:'La verdadera grandeza del líder no viene de su cargo, sino de depender de Dios y obedecer sus principios.',ref:'Christian Leadership, pp. 13-14'},
  {kind:'Elena G. de White',text:'Un dirigente necesita permanecer como aprendiz y pedir un corazón capaz de discernir entre el bien y el mal.',ref:'Christian Leadership, pp. 13-14'},
  {kind:'Elena G. de White',text:'Más oración y más humildad en quienes llevan responsabilidades permiten una dirección más claramente guiada por Dios.',ref:'Christian Leadership, Committee Meetings'},
  {kind:'Elena G. de White',text:'El liderazgo cristiano no consiste en controlar vidas, sino en formar personas capaces de servir y asumir responsabilidad.',ref:'Christian Leadership, Training Leaders / Individuality'},
  {kind:'Elena G. de White',text:'Los jóvenes crecen como líderes cuando reciben oportunidades reales para participar, expresarse y asumir responsabilidades progresivamente.',ref:'Testimonies on Sabbath-School Work, p. 48'},
  {kind:'Elena G. de White',text:'La oración sencilla y ferviente debe acompañar a quienes enseñan y conducen a otros.',ref:'Prayer, pp. 248-249'},
  {kind:'Elena G. de White',text:'Quienes sirven juntos necesitan también animarse, apoyarse y orar unos por otros en momentos de carga.',ref:'Prayer, p. 249'},
  {kind:'Elena G. de White',text:'El liderazgo se fortalece cuando el trabajo se realiza con orden y sin gastar energías en asuntos menores.',ref:'Christian Leadership, Committee Meetings'},
  {kind:'Servicio',text:'Hoy puedes dirigir menos con órdenes y más con ejemplo, escucha y servicio.',ref:'Principio de liderazgo cristiano'},
  {kind:'Oración',text:'Antes de tomar una decisión importante, reserva un momento para pedir sabiduría, humildad y claridad.',ref:'Santiago 1:5'},
  {kind:'Misión',text:'Un líder forma nuevos líderes cuando confía responsabilidades, acompaña y permite que otros crezcan.',ref:'2 Timoteo 2:2'}
];
let sglLoginMessageIndex = -1;
function sglNextLoginMessage(){
  if(window.SGL_INSPIRATION?.pick){const item=window.SGL_INSPIRATION.pick('login-app');return {kind:item.kind,group:item.group,text:item.text,ref:item.source,url:item.url||''};}
  if(!SGL_LOGIN_MESSAGES.length)return {kind:'Liderazgo',text:'Liderar es servir, guiar y dar ejemplo.',ref:''};
  let next=0;try{const a=new Uint32Array(1);crypto.getRandomValues(a);next=a[0]%SGL_LOGIN_MESSAGES.length;}catch{next=Math.floor(Math.random()*SGL_LOGIN_MESSAGES.length);}
  if(SGL_LOGIN_MESSAGES.length>1&&next===sglLoginMessageIndex)next=(next+1)%SGL_LOGIN_MESSAGES.length;
  sglLoginMessageIndex=next;return SGL_LOGIN_MESSAGES[next];
}
function sgl5065SourceLink(item){return item?.url?`<a class="sgl5065-inspiration-link" href="${esc(item.url)}" target="_blank" rel="noopener noreferrer">Libros de Elena G. de White ↗</a>`:'';}
function sgl5066DisplaySource(item){
  const source=String(item?.source??item?.ref??'');
  if(item?.group==='famous'||item?.kind==='Personajes notables')return `${source}${source?' · ':''}Personajes notables`;
  if(item?.group==='egw'||item?.kind==='Elena G. de White')return source.startsWith('Elena G. de White')?source:`Elena G. de White${source?' · ':''}${source}`;
  return source;
}
function leadershipAuthHero(subtitle='Ministerio de Conquistadores') {
  const item=sglNextLoginMessage();
  return `<div class="auth-hero"><div class="brand"><img src="${appLogo()}"><div><strong>Sistema de Gestión de Música</strong><small>${esc(subtitle)}</small></div></div><h1>Liderar es servir, orar y dar ejemplo.</h1><div class="leadership-verses"><p class="leadership-verse login-rotating-message">“${esc(item.text)}” <strong>${esc(sgl5066DisplaySource(item))}</strong>${sgl5065SourceLink(item)}</p></div></div>`;
}

function renderRemoteLogin(message='') {
  document.documentElement.dataset.theme='classic_blue';
  document.body.innerHTML = `<section class="screen auth">${leadershipAuthHero('Acceso remoto seguro')}<div class="auth-panel"><form class="auth-card" onsubmit="remoteLogin(event)"><div class="eyebrow">Servidor oficial</div><h2>Iniciar sesión</h2>${authConnectionNotice(message)}<label class="field-group">Usuario o correo<input id="remote_login_user" class="field" required autocomplete="username"></label><label class="field-group">Contraseña<input id="remote_login_password" type="password" class="field" required autocomplete="current-password"></label><div id="remote_login_error" class="error"></div><button id="remote_login_button" class="btn primary block" type="submit">Ingresar</button><div class="notice" style="margin-top:14px"><b>Privacidad:</b> el servidor entrega los datos autorizados según el rol. Un Administrador remoto recibe acceso completo para gestionar desde la web o el celular.</div></form></div></section>`;
  decorateAuth();
  sglApplyOfflineMode(!navigator.onLine);
}

async function sglRefreshRemoteFilesAfterLogin(expectedToken='') {
  if(!REMOTE_WEB||!expectedToken)return;
  // 5.0.48: no descargar de golpe todas las fotos/PDF al iniciar sesión móvil.
  // El servidor entrega solo el manifiesto y cada evidencia se obtiene por fileId
  // cuando el informe realmente la necesita. Así el acceso no queda bloqueado por
  // decenas de MB de imágenes ni se vacía la caché local antes de tiempo.
  const sessionStillActive=()=>database?.settings?.serverToken===expectedToken&&sessionStorage.getItem('musica_session_user');
  if(!sessionStillActive())return;
  try {
    const response=await fetch(`${REMOTE_ORIGIN}/api/remote/files?manifest=1`,{method:'GET',headers:{'Authorization':`Bearer ${expectedToken}`,'Cache-Control':'no-cache','Pragma':'no-cache'},cache:'no-store'});
    if(!response.ok||!sessionStillActive())return;
    const payload=await response.json().catch(()=>({}));
    if(!payload.ok||!sessionStillActive())return;
    const incoming=Array.isArray(payload.files)?payload.files:[];
    window.__SGL_REMOTE_FILE_MANIFEST__=new Map(incoming.map(file=>[String(file?.id||''),file]));
    if(sessionStillActive())window.dispatchEvent(new CustomEvent('sgl-remote-files-ready',{detail:{count:incoming.length,manifestOnly:true}}));
  } catch {}
}

async function remoteLogin(event) {
  event.preventDefault();
  if (!navigator.onLine && !REMOTE_WEB) return setText('remote_login_error','Necesitas conexión para el primer acceso. Si ya habías ingresado antes, usa la pantalla de acceso offline guardada en este dispositivo.');
  const identity=normalize(value('remote_login_user')), password=value('remote_login_password'), button=document.getElementById('remote_login_button');
  if(button){button.disabled=true;button.textContent='Verificando…';}
  try {
    const response=await fetch(`${REMOTE_ORIGIN}/api/remote/login`,{method:'POST',headers:{'Content-Type':'application/json','Cache-Control':'no-cache','Pragma':'no-cache'},body:JSON.stringify({identity,password,deviceId:database.installationId||newId('install')}),cache:'no-store'});
    const payload=await response.json().catch(()=>({}));
    if(!response.ok) throw new Error(payload.error||`HTTP ${response.status}`);
    const localInstallationId=database.installationId||newId('install');
    const priorRestoreGeneration=Math.max(1,Number(database?.restoreGeneration||1)||1);
    const priorRestoreId=String(database?.restoreId||'');
    const priorRestoredAt=String(database?.restoredAt||'');
    const priorServerVersion=Number(database?.serverVersion||0)||0;
    const priorControlRevision=Number(database?.settings?.serverControlRevision||0)||0;
    database=payload.database||blankDatabase();
    database.installationId=localInstallationId;
    database.restoreGeneration=Math.max(priorRestoreGeneration,Number(payload.restoreGeneration||payload.control?.restoreGeneration||database.restoreGeneration||1)||1);
    database.restoreId=String(payload.restoreId||payload.control?.restoreId||priorRestoreId||database.restoreId||'');
    database.restoredAt=String(payload.restoredAt||payload.control?.issuedAt||priorRestoredAt||database.restoredAt||'');
    database.restorePending=false;
    database.serverVersion=Math.max(priorServerVersion,Number(payload.serverVersion||database.serverVersion||0)||0);
    database.settings ||= {};
    database.settings.serverUrl=REMOTE_ORIGIN;
    database.settings.serverToken=payload.sessionToken||'';
    database.settings.autoSync=true;
    database.settings.remoteAccess=true;
    database.settings.remoteUserId=payload.userId||'';
    database.settings.serverControlRevision=Math.max(priorControlRevision,Number(payload.controlRevision||payload.control?.revision||0)||0);
    try{sessionStorage.setItem('sgl_control_seen_revision',String(database.settings.serverControlRevision||0));}catch{}
    migrateDatabase();
    database.sync={state:'ok',lastSyncedAt:isoNow(),message:'Acceso remoto autenticado',pendingProposal:null,needsUpload:false};
    await kvSet('database',database);
    sessionStorage.setItem('musica_session_user',payload.userId);
    // El primer acceso autorizado deja explícitamente preparada la copia offline.
    sglPrepareOfflineShell().catch(()=>{});
    renderShell();
    // Las evidencias pesadas se descargan después de mostrar el menú. Así el login no espera fotos/PDF.
    sglRefreshRemoteFilesAfterLogin(database.settings.serverToken).catch(()=>{});
  } catch(error) {
    const message=error.message||'No se pudo iniciar sesión.';
    setText('remote_login_error',message);
    sglApplyOfflineMode(true,'Servidor no disponible o conexión interrumpida. SGL queda en modo offline hasta recuperar conexión.');
    if(button){button.disabled=false;button.textContent='Ingresar';}
  }
}

function renderBootstrap() {
  document.documentElement.dataset.theme='classic_blue';
  document.body.innerHTML = `<section class="screen auth"><div class="auth-hero"><div class="brand"><img src="${appLogo()}"><div><strong>Sistema de Gestión de Música</strong><small>Configuración oficial</small></div></div><h1>Primero, crea la cuenta del Administrador.</h1><p>Solo el Administrador podrá crear Tutores y Aspirantes. No se cargarán usuarios ni avances ficticios.</p></div><div class="auth-panel"><form class="auth-card" onsubmit="bootstrapAdmin(event)"><div class="eyebrow">Primer inicio · versión ${APP_VERSION}</div><h2>Administrador principal</h2><p>Estos datos se guardan en el dispositivo y podrán sincronizarse con el servidor oficial.</p><label class="field-group">Nombres y Apellidos Completos<input id="boot_name" class="field" required autocomplete="name"></label><label class="field-group">Correo electrónico<input id="boot_email" type="email" class="field" required autocomplete="email"></label><label class="field-group">Nombre de usuario<input id="boot_user" class="field" required autocomplete="username"></label><label class="field-group">Contraseña<input id="boot_password" type="password" minlength="8" class="field" required autocomplete="new-password"></label><div id="boot_error" class="error"></div><button class="btn primary block" type="submit">Crear Administrador</button></form></div></section>`;
  decorateAuth();
}

async function bootstrapAdmin(event) {
  event.preventDefault();
  if (database.users.some(user => user.role === 'admin')) return;
  const fullName = value('boot_name').trim();
  const email = normalize(value('boot_email'));
  const username = normalize(value('boot_user'));
  const password = value('boot_password');
  if (!fullName || !email || !username || password.length < 8) return setText('boot_error', 'Completa todos los campos; la contraseña requiere al menos 8 caracteres.');
  const recoveryKeys=[makeRecoveryKey(),makeRecoveryKey()];
  const user = {id:newId('usr'), fullName, email, username, role:'admin', active:true, accountState:'active', theme:'classic_blue', createdAt:isoNow(), password:await passwordRecord(password), recoveryKeys:[await passwordRecord(recoveryKeys[0]),await passwordRecord(recoveryKeys[1])], photo:''};
  database.users.push(user);
  await persist(true, {action:'Administrador inicial creado', entity:user.id, detail:email});
  document.body.innerHTML=`<section class="screen auth"><div class="auth-panel" style="width:min(700px,95vw)"><div class="auth-card"><div class="eyebrow">Llaves maestras del Administrador</div><h2>Guárdalas ahora</h2><p>Son de un solo uso y permiten recuperar la cuenta sin correo. La aplicación guarda únicamente su huella cifrada y no podrá volver a mostrarlas.</p><div class="recovery-keys"><code>${esc(recoveryKeys[0])}</code><code>${esc(recoveryKeys[1])}</code></div><button class="btn ghost block" onclick="downloadRecoveryKeys('${esc(recoveryKeys[0])}','${esc(recoveryKeys[1])}')">Descargar llaves</button><button class="btn primary block" style="margin-top:10px" onclick="sessionStorage.setItem('musica_session_user','${user.id}');localStorage.setItem('musica_session_user','${user.id}');renderShell()">Ya las guardé · continuar</button></div></div></section>`;
}

function makeRecoveryKey() { const bytes=crypto.getRandomValues(new Uint8Array(12)); return `CA-${[...bytes].map(byte=>byte.toString(16).padStart(2,'0')).join('').toUpperCase().match(/.{1,4}/g).join('-')}`; }
function downloadRecoveryKeys(first,second) { downloadText('LLAVES_MAESTRAS_SGL.txt',`SISTEMA DE GESTIÓN DE LIDERAZGO · LLAVES DEL ADMINISTRADOR\n\nLlave 1: ${first}\nLlave 2: ${second}\n\nCada llave es de un solo uso. Guarde este archivo fuera del equipo.\n`); }

function renderLogin(message='') {
  sessionStorage.removeItem('musica_session_user');
  localStorage.removeItem('musica_session_user');
  document.documentElement.dataset.theme='classic_blue';
  const remoteHint = REMOTE_WEB ? `<div class="notice" style="margin-top:12px"><b>Modo offline:</b> si este dispositivo ya fue autorizado, puedes ingresar con tu contraseña aunque el servidor esté temporalmente apagado.</div>` : '';
  const externalLogin = REMOTE_WEB ? `<button class="btn ghost block" style="margin-top:10px" type="button" onclick="renderRemoteLogin()">Actualizar acceso con el servidor</button><p class="muted" style="font-size:11px;margin-top:8px">Úsalo si cambiaste de cuenta, contraseña o permisos. No requiere aprobación del Administrador.</p>` : `<button class="btn link block" type="button" onclick="renderPasswordRecovery()">Olvidé mi contraseña</button><div class="divider">o</div><button class="btn oauth block" type="button" onclick="googleLogin()">Continuar con Google y Drive</button><p class="muted" style="font-size:11px;margin-top:13px">Google no crea cuentas nuevas: el correo debe haber sido habilitado previamente por el Administrador.</p>`;
  const rememberLocal = REMOTE_WEB ? '' : `<label style="display:flex;align-items:center;gap:8px;margin:2px 0 14px;font-size:13px"><input id="login_remember" type="checkbox" checked> Mantener sesión iniciada en este equipo</label>`;
  document.body.innerHTML = `<section class="screen auth">${leadershipAuthHero()}<div class="auth-panel"><form class="auth-card" onsubmit="login(event)"><div class="eyebrow">Acceso seguro</div><h2>Iniciar sesión</h2><p>Usa la cuenta creada por el Administrador.</p>${authConnectionNotice(message)}<label class="field-group">Usuario o correo<input id="login_user" class="field" required autocomplete="username"></label><label class="field-group">Contraseña<input id="login_password" type="password" class="field" required autocomplete="current-password"></label>${rememberLocal}<div id="login_error" class="error"></div><button class="btn primary block" type="submit">Ingresar</button>${remoteHint}${externalLogin}</form></div></section>`;
  decorateAuth();
  sglApplyOfflineMode(!navigator.onLine);
}

function decorateAuth(){const hero=document.querySelector('.auth-hero');if(hero)hero.insertAdjacentHTML('afterbegin',`<img class="auth-watermark" src="${appLogo()}" alt="">`);}

function renderPasswordRecovery(message='') {
  document.body.innerHTML=`<section class="screen auth"><div class="auth-panel" style="width:min(620px,95vw)"><form class="auth-card" onsubmit="recoverPassword(event)"><div class="eyebrow">Recuperar contraseña</div><h2>Recuperación segura</h2><p>El Administrador puede usar una de sus dos llaves maestras. Los demás usuarios deben solicitar una contraseña temporal al Administrador o usar Google con su correo habilitado.</p>${message?`<div class="notice warn">${esc(message)}</div>`:''}<label class="field-group">Usuario o correo<input id="recover_identity" class="field" required></label><label class="field-group">Llave maestra del Administrador<input id="recover_key" class="field" autocomplete="off"></label><label class="field-group">Nueva contraseña<input id="recover_password" type="password" minlength="8" class="field" required></label><div id="recover_error" class="error"></div><button class="btn primary block">Cambiar contraseña</button><button class="btn ghost block" type="button" onclick="renderLogin()">Volver</button></form></div></section>`;
}

async function recoverPassword(event) {
  event.preventDefault(); const identity=normalize(value('recover_identity')), key=value('recover_key').trim(), next=value('recover_password');
  const user=database.users.find(item=>normalize(item.username)===identity||normalize(item.email)===identity);
  if(!user||user.role!=='admin'||!key) return setText('recover_error','Para esta cuenta solicita una contraseña temporal al Administrador.');
  const index=await firstMatchingRecoveryKey(user,key); if(index<0)return setText('recover_error','Llave maestra incorrecta o ya utilizada.');
  if(next.length<8)return setText('recover_error','La nueva contraseña requiere al menos 8 caracteres.');
  const replacement=makeRecoveryKey(); user.password=await passwordRecord(next); user.recoveryKeys[index]=await passwordRecord(replacement);
  await persist(true,{action:'Administrador recuperado con llave maestra',entity:user.id,detail:`Llave ${index+1} reemplazada`});
  document.body.innerHTML=`<section class="screen auth"><div class="auth-panel"><div class="auth-card"><h2>Contraseña recuperada</h2><p>La llave usada quedó anulada. Guarda su reemplazo:</p><div class="recovery-keys"><code>${esc(replacement)}</code></div><button class="btn primary block" onclick="renderLogin('Contraseña actualizada. Ya puedes ingresar.')">Continuar</button></div></div></section>`;
}

async function firstMatchingRecoveryKey(user,key) { for(let index=0;index<(user.recoveryKeys||[]).length;index+=1)if(await verifyPassword(key,user.recoveryKeys[index]))return index; return -1; }

async function login(event) {
  event.preventDefault();
  const identity = normalize(value('login_user'));
  const password = value('login_password');
  const user = database.users.find(candidate => (candidate.accountState||'active') === 'active' && (normalize(candidate.username) === identity || normalize(candidate.email) === identity));
  if (!user || !(await verifyPassword(password, user.password))) return setText('login_error', 'Usuario o contraseña incorrectos.');
  if(window.SGL593Allows&&!window.SGL593Allows(database,user,true))return setText('login_error','Tu cuenta no tiene acceso a este portal. Usa el enlace del ministerio asignado.');
  sessionStorage.setItem('musica_session_user', user.id);
  if(!REMOTE_WEB){try{if(document.getElementById('login_remember')?.checked!==false)localStorage.setItem('musica_session_user',user.id);else localStorage.removeItem('musica_session_user');}catch{}}
  addAudit('Inicio de sesión', user.id, 'Acceso local');
  await persist(false);
  renderShell();
}

async function logout() {
  const remoteToken=REMOTE_WEB ? (database?.settings?.serverToken||'') : '';
  addAudit('Cierre de sesión', currentUser()?.id || '', '');
  await persist(false);
  if(REMOTE_WEB&&remoteToken&&navigator.onLine){try{await fetch(`${REMOTE_ORIGIN}/api/remote/logout`,{method:'POST',headers:{'Authorization':`Bearer ${remoteToken}`}});}catch{}}
  stopRemoteHeartbeat();
  sessionStorage.removeItem('musica_session_user');
  localStorage.removeItem('musica_session_user');
  activeAspirantId = null;
  activeClassId = '';
  activeClassSection = 'dashboard';
  currentPage = 'dashboard';
  sglHistoryReady = false;
  try{history.replaceState({sglLogin:true},'',location.href);}catch{}
  renderLogin();
}

async function remoteHeartbeat(){
  if(!REMOTE_WEB||!database?.settings?.serverToken||!currentUser())return;
  try{await fetch(`${REMOTE_ORIGIN}/api/remote/heartbeat`,{method:'POST',headers:{'Authorization':`Bearer ${database.settings.serverToken}`,'Cache-Control':'no-cache','Pragma':'no-cache'},cache:'no-store'});}catch{}
}
function startRemoteHeartbeat(){
  if(!REMOTE_WEB)return;
  clearInterval(remoteHeartbeatTimer);
  remoteHeartbeat();
  remoteHeartbeatTimer=setInterval(remoteHeartbeat,20000);
}
function stopRemoteHeartbeat(){clearInterval(remoteHeartbeatTimer);remoteHeartbeatTimer=null;}

async function closeApplication() {
  const snapshot=await writeNativeSnapshot();
  if(snapshot?.ok)try{await nativeRequest('/data/prepare-clean',{});}catch{}
  fetch('/api/local/shutdown',{method:'POST'}).catch(()=>{}).finally(()=>{ try{window.close();}catch{} });
}

async function passwordRecord(password) {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const derived = await deriveKeyBytes(password, salt, 150000, 32);
  return {algorithm:'PBKDF2-SHA256', iterations:150000, salt:toBase64(salt), hash:toBase64(derived)};
}

async function verifyPassword(password, record) {
  if (!record?.salt || !record?.hash) return false;
  const salt = fromBase64(record.salt);
  const derived = await deriveKeyBytes(password, salt, record.iterations || 150000, 32);
  return timingSafeEqual(derived, fromBase64(record.hash));
}

async function deriveKeyBytes(password, salt, iterations, length) {
  const material = await crypto.subtle.importKey('raw', new TextEncoder().encode(password), 'PBKDF2', false, ['deriveBits']);
  const bits = await crypto.subtle.deriveBits({name:'PBKDF2', hash:'SHA-256', salt, iterations}, material, length * 8);
  return new Uint8Array(bits);
}

function timingSafeEqual(a, b) { if (a.length !== b.length) return false; let diff = 0; for (let i = 0; i < a.length; i += 1) diff |= a[i] ^ b[i]; return diff === 0; }
function toBase64(bytes) { let binary=''; bytes.forEach(byte => binary += String.fromCharCode(byte)); return btoa(binary); }
function fromBase64(value) { const binary=atob(value); return Uint8Array.from(binary, char => char.charCodeAt(0)); }

function renderShell() {
  const user = currentUser();
  if (!user) return renderLogin();
  applyTheme();
  if (['aspirant','pastor_aspirant'].includes(user.role)) activeAspirantId = user.id;
  if (!activeAspirantId && ['admin','tutor','lider_musica'].includes(user.role)) activeAspirantId = accessibleAspirants(user).at(0)?.id || null;
  document.body.innerHTML = `<section class="shell"><aside class="sidebar"><div class="side-brand"><img src="${appLogo()}"><div><strong>Sistema de Gestión de Música</strong><small>Versión ${APP_VERSION}</small></div></div><nav id="nav" class="nav"></nav><div class="side-user"><div class="user-row"><div class="avatar" id="side_avatar">${avatarMarkup(user)}</div><div class="user-meta"><strong>${esc(user.fullName)}</strong><small>${ROLE_LABELS[user.rawRole624||user.role]}</small></div><button class="logout" onclick="logout()" title="Cerrar sesión" aria-label="Cerrar sesión">CERRAR SESIÓN</button></div></div></aside><main class="main"><header class="topbar"><div class="music626-top-context" hidden><div class="eyebrow" id="top_eyebrow">Sistema de Gestión de Música</div><h2 id="top_title">Panel</h2></div><div class="spacer"></div>${user.role==='admin'?'<button id="presence_pill" class="presence-pill" onclick="v3ShowConnectedUsers()" title="Usuarios conectados ahora"><span class="presence-dot"></span><span id="presence_pill_text">0 conectados</span></button>':''}<button id="sync_pill" class="sync-pill sync-local" onclick="syncNow()">Guardado local</button></header><div id="content" class="content"></div></main></section><button class="mobile-drawer-tab" type="button" onclick="toggleMobileMenu()" aria-label="Desplegar menú" aria-expanded="false">›</button><button class="floating-refresh" onclick="refreshCurrentView()" title="Actualizar datos y refrescar esta pantalla" aria-label="Actualizar">↻</button><div id="modal_root"></div><div id="toast_root"></div>`;
  renderNav();
  setupMobileDrawer();
  sglEnsureHistoryRoot(currentPage || 'dashboard');
  go(currentPage || 'dashboard',{history:false});
  renderTopSync();
  if(REMOTE_WEB) startRemoteHeartbeat(); else stopRemoteHeartbeat();
  // Revalida el inicio del sábado después de que el shell (incluido toast_root)
  // ya existe. El sondeo global también cubre el momento de la puesta de sol
  // cuando la aplicación permanece abierta.
  setTimeout(()=>{sglSabbathPrimeCoordinates().catch(()=>{});sglSabbathPoll().catch(()=>{});},80);
  if(!REMOTE_WEB&&user.role==='admin'&&!(user.recoveryKeys||[]).length)setTimeout(()=>createAdminRecoveryKeys(user.id),200);
  if(!taskReminderShown)setTimeout(showAssignedTaskReminder,350);
}

async function sglRefreshWithoutNavigation(){
  try{
    if(typeof refreshCurrentView==='function'&&currentUser()){await refreshCurrentView();return;}
    if(REMOTE_WEB&&typeof remoteHeartbeat==='function')await remoteHeartbeat();
    try{window.sgl509UpdateOfflineIndicator?.();}catch{}
    toast('Estado actualizado sin recargar la pagina.');
  }catch{toast('La vista local permanece disponible.');}
}
window.sglRefreshWithoutNavigation=sglRefreshWithoutNavigation;

async function refreshCurrentView(){
  const navigationEpoch=Number(window.__sgl555NavigationEpoch||0),page=currentPage,classId=activeClassId,classSection=activeClassSection,latest=await kvGet('database');
  if(REMOTE_WEB){
    // El boton azul revalida la release y la preparacion offline sin recargar el
    // documento; asi no entrega la navegacion al error negro del navegador.
    try{const reg=await navigator.serviceWorker?.getRegistration?.();await reg?.update?.();}catch{}
    try{await sglPrepareOfflineShell();}catch{}
  }
  if(latest&&Number(latest.revision||0)>=Number(database.revision||0)){database=latest;migrateDatabase();}
  renderShell();
  if(navigationEpoch===Number(window.__sgl555NavigationEpoch||0)&&page==='classes'&&classId&&accessibleClasses().some(item=>item.id===classId)){activeClassId=classId;activeClassSection=classSection;renderClassWorkspace();}
  try{await window.sgl509UpdateOfflineIndicator?.();}catch{}
  toast('Vista actualizada.');
}

async function createAdminRecoveryKeys(userId){if(REMOTE_WEB)return;const user=userById(userId);if(!user||user.recoveryKeys?.length)return;const keys=[makeRecoveryKey(),makeRecoveryKey()];user.recoveryKeys=[await passwordRecord(keys[0]),await passwordRecord(keys[1])];await persist(true,{action:'Llaves maestras creadas',entity:user.id,detail:'2 llaves de un solo uso'});showModal(`<div class="modal small"><div class="modal-head"><div><h3>Llaves maestras del Administrador</h3><p>Guárdalas antes de continuar.</p></div></div><div class="modal-body"><div class="recovery-keys"><code>${esc(keys[0])}</code><code>${esc(keys[1])}</code></div><div class="notice warn">No volverán a mostrarse. Cada una se usa una sola vez y genera un reemplazo.</div></div><div class="modal-foot"><button class="btn ghost" onclick="downloadRecoveryKeys('${esc(keys[0])}','${esc(keys[1])}')">Descargar</button><button class="btn primary" onclick="closeModal()">Ya las guardé</button></div></div>`);}

function mobileDrawerPortrait(){return window.matchMedia('(max-width:760px) and (orientation:portrait)').matches;}
function setMobileDrawerState(open){
  if(!document.body)return;
  if(!mobileDrawerPortrait())open=false;
  document.body.classList.toggle('mobile-menu-open',!!open);
  const tab=document.querySelector('.mobile-drawer-tab');
  if(tab){tab.textContent=open?'‹':'›';tab.setAttribute('aria-expanded',open?'true':'false');tab.setAttribute('aria-label',open?'Cerrar menú':'Desplegar menú');}
}
function openMobileMenu(){setMobileDrawerState(true);}
function closeMobileMenu(){setMobileDrawerState(false);}
function toggleMobileMenu(){setMobileDrawerState(!document.body.classList.contains('mobile-menu-open'));}
function setupMobileDrawer(){
  if(window.__sglMobileDrawerBound)return;
  window.__sglMobileDrawerBound=true;
  let startX=0,startY=0,tracking=false;
  document.addEventListener('touchstart',event=>{
    if(!mobileDrawerPortrait()||event.touches.length!==1)return;
    const touch=event.touches[0];startX=touch.clientX;startY=touch.clientY;
    tracking=startX<=50||document.body.classList.contains('mobile-menu-open');
  },{passive:true});
  document.addEventListener('touchend',event=>{
    if(!tracking||!mobileDrawerPortrait()||!event.changedTouches.length){tracking=false;return;}
    const touch=event.changedTouches[0],dx=touch.clientX-startX,dy=touch.clientY-startY;tracking=false;
    if(Math.abs(dx)<55||Math.abs(dx)<Math.abs(dy)*1.15)return;
    if(dx>0&&startX<=50)openMobileMenu();
    else if(dx<0&&document.body.classList.contains('mobile-menu-open'))closeMobileMenu();
  },{passive:true});
  document.addEventListener('click',event=>{
    if(!document.body.classList.contains('mobile-menu-open'))return;
    if(event.target.closest('.sidebar,.mobile-drawer-tab'))return;
    closeMobileMenu();
  });
  window.addEventListener('orientationchange',()=>setTimeout(closeMobileMenu,120));
  window.addEventListener('resize',()=>{if(!mobileDrawerPortrait())closeMobileMenu();});
}

function avatarMarkup(user) { return user?.photo ? `<img src="${user.photo}" alt="">` : esc(initials(user?.fullName)); }

function navItems(user) {
  if (user.role === 'admin') return [
    ['dashboard','▦','Inicio'], ['tutors','♙','Tutores'], ['aspirants','◎','Aspirantes'], ['tasks','◷','Tareas'], ['reviews','✓','Revisiones'], ['indications','✎','Indicaciones'], ['materials','▣','Materiales'], ['classes','◆','Clases'],
    ['clubs','◇','Clubes'], ['leadership','🏆','Liderazgo en Acción'], ['drive','△','Google Drive'], ['profile','○','Mi perfil'], ['settings','⚙','Configuración'], ['users','☷','Usuarios']
  ];
  if (user.role === 'tutor') return [
    ['dashboard','▦','Resumen'], ['aspirants','♙','Mis Aspirantes'], ['tasks','◷','Tareas'], ['materials','▣','Materiales'], ['reviews','✓','Revisiones'],
    ['requirements','☷','Expediente'], ['leadership','🏆','Liderazgo en Acción'], ['drive','△','Google Drive'], ['profile','○','Mi perfil'], ['settings','⚙','Configuración']
  ];
  return [
    ['dashboard','▦','Inicio'], ['requirements','☷','Mis requisitos'], ['reports','□','Mis informes'], ['tasks','◷','Mis tareas'], ['bible','▤','Lectura bíblica'],
    ...(ClassPolicy.canAccessClasses(user,database.classes)?[['classes','◆','Mis clases']]:[]),
    ['materials','▣','Materiales'], ['leadership','🏆','Liderazgo en Acción'], ['drive','△','Google Drive'], ['profile','○','Mi perfil'], ['settings','⚙','Configuración']
  ];
}

function renderNav() {
  const user = currentUser();
  const reviewCount = pendingReviews(user).length;
  document.getElementById('nav').innerHTML = navItems(user).map(([page, icon, label]) => `<button data-page="${page}" class="${currentPage === page ? 'active' : ''}${page==='users'?' nav-bottom':''}" onclick="go('${page}')"><b>${icon}</b><span>${label}</span>${page === 'reviews' && reviewCount ? `<i class="count">${reviewCount}</i>` : ''}${page === 'conflicts' && database.conflicts.length ? `<i class="count">${database.conflicts.length}</i>` : ''}</button>`).join('');
}

function sglHistoryPageState(page=currentPage,depth=0){return {sgl:true,kind:'page',page:page||'dashboard',depth:Math.max(0,Number(depth)||0)};}
function sglEnsureHistoryRoot(page='dashboard'){
  if(sglHistoryReady&&history.state?.sgl)return;
  try{history.replaceState(sglHistoryPageState(page,0),'',location.href);sglHistoryReady=true;}catch{}
}
function sglPushHistoryState(state){
  if(sglHistoryRestoring)return;
  sglEnsureHistoryRoot(currentPage||'dashboard');
  const base=history.state?.sgl?history.state:sglHistoryPageState(currentPage,0),depth=Math.max(0,Number(base.depth)||0)+1;
  try{history.pushState({...state,sgl:true,depth},'',location.href);}catch{}
}
function sglPushTransientHistory(kind){
  if(sglHistoryRestoring)return;
  sglEnsureHistoryRoot(currentPage||'dashboard');
  const base=history.state?.sgl?structuredClone(history.state):sglHistoryPageState(currentPage,0);
  if(base.kind===kind)return;
  sglPushHistoryState({kind,page:currentPage,base});
}
function sglDiscardTransientHistory(){
  const state=history.state;
  if(!state?.sgl||!['modal','photo'].includes(state.kind))return;
  const base=state.base?.sgl?state.base:sglHistoryPageState(currentPage,Math.max(0,(Number(state.depth)||1)-1));
  try{history.replaceState(base,'',location.href);}catch{}
}
function sglNavigateHistoryBack(){
  const state=history.state;
  if(state?.sgl&&Number(state.depth||0)>0){history.back();return true;}
  return false;
}
async function sglRestoreHistoryState(state){
  if(!state?.sgl)return;
  sglHistoryRestoring=true;
  try{
    if(state.kind==='class'){
      if(document.querySelector('#modal_root .modal-bg')&&await closeModal(true)===false)return;
      currentPage='classes';activeClassId=state.classId||'';activeClassSection=state.section||'dashboard';
      renderNav();
      if(activeClassId&&accessibleClasses().some(item=>String(item.id)===String(activeClassId)))renderClassWorkspace();else{activeClassId='';renderClasses();}
      return;
    }
    if(state.kind==='page'){
      if(document.querySelector('#modal_root .modal-bg')&&await closeModal(true)===false)return;
      if(state.page==='classes')activeClassId='';
      go(state.page||'dashboard',{history:false});
    }
  } finally {sglHistoryRestoring=false;}
}
window.addEventListener('popstate',async event=>{
  if(!currentUser()){
    sglHistoryReady=false;
    try{history.replaceState({sglLogin:true},'',location.href);}catch{}
    renderLogin();
    return;
  }
  const state=event.state,hadViewer=!!document.getElementById('sgl_image_viewer'),hadModal=!!document.querySelector('#modal_root .modal-bg');
  let cancelled=false;sglHistoryRestoring=true;
  try{
    if(hadViewer)closeImageViewer(true);
    if(hadModal&&state?.kind!=='modal')cancelled=await closeModal(true)===false;
  }finally{sglHistoryRestoring=false;}
  if(cancelled){sglPushTransientHistory('modal');return;}
  if(state?.sgl&&state.kind!=='modal'&&state.kind!=='photo')await sglRestoreHistoryState(state);
});

function go(page,options={}) {
  closeMobileMenu();
  const user = currentUser();
  const pages = {
    dashboard:renderDashboard, classes:renderClasses, users:renderUsers, tutors:renderTutors, clubs:renderClubs, aspirants:renderAspirants,
    reviews:renderReviews, tasks:renderTasks, materials:renderMaterials, indications:renderIndications, requirements:renderRequirements, bible:renderBible,
    reports:renderReports, drive:renderDrive, backup:renderBackup, profile:renderProfile,
    settings:renderSettings, conflicts:renderConflicts, leadership:renderLeadership
  };
  if (!pages[page]) page = 'dashboard';
  if(page==='classes'&&!ClassPolicy.canAccessClasses(user,database.classes))page='dashboard';
  if ((page === 'users' || page === 'clubs' || page === 'conflicts' || page === 'indications') && user.role !== 'admin') page='dashboard';
  const previous=currentPage;
  if(history.state?.sgl&&['modal','photo'].includes(history.state.kind))sglDiscardTransientHistory();
  currentPage = page;
  closeModal(true);
  if(options.history!==false&&!sglHistoryRestoring&&previous!==page){
    if(options.replace){const depth=Math.max(0,Number(history.state?.depth)||0);try{history.replaceState(sglHistoryPageState(page,depth),'',location.href);}catch{}}
    else sglPushHistoryState({kind:'page',page});
  }
  renderNav();
  pages[page]();
}

window.sglAndroidBackAction=function(){
  if(document.getElementById('sgl_image_viewer')){closeImageViewer();return 'handled';}
  if(document.querySelector('#modal_root .modal-bg')){closeModal();return 'handled';}
  if(document.body.classList.contains('mobile-menu-open')){closeMobileMenu();return 'handled';}
  if(document.querySelector('.class-workspace')){
    if(!sglNavigateHistoryBack()){activeClassId='';activeClassSection='dashboard';renderClasses();}
    return 'handled';
  }
  if(currentPage!=='dashboard'){
    if(!sglNavigateHistoryBack())go('dashboard',{replace:true});
    return 'handled';
  }
  return 'exit';
};

function setTop(title, eyebrow='Clase Agrupada') {
  setText('top_title', title);
  setText('top_eyebrow', eyebrow);
}

function accessibleAspirants(user=currentUser()) {
  const aspirants = database.users.filter(candidate => ['aspirant','pastor_aspirant'].includes(candidate.role) && (candidate.accountState||'active') === 'active');
  if (!user) return [];
  if (user.role === 'admin' || user.role === 'regional') return aspirants;
  if (user.role === 'tutor') return aspirants.filter(candidate => sgl545HasTutor(candidate,user.id));
  return aspirants.filter(candidate => candidate.id === user.id);
}

function targetAspirant() {
  const user = currentUser();
  if(!user)return null;
  if (['aspirant','pastor_aspirant'].includes(user.role)) return user;
  const allowed = accessibleAspirants(user);
  if (!allowed.some(candidate => candidate.id === activeAspirantId)) activeAspirantId = allowed.at(0)?.id || null;
  return userById(activeAspirantId);
}

function ensureProgress(userId) {
  if (!database.progress[userId]) database.progress[userId] = {status:{}, submissions:{}, readings:{}, updatedAt:isoNow()};
  const progress = database.progress[userId];
  progress.status ||= {};
  progress.submissions ||= {};
  progress.readings ||= {};
  for (const requirement of flatRequirements) progress.status[requirement.id] ||= 'pending';
  return progress;
}

function submissionFor(userId, requirementId) {
  const progress = ensureProgress(userId);
  progress.submissions[requirementId] ||= {text:'', richText:'', completionDate:'', selections:[], attachmentIds:[], feedback:'', updatedAt:isoNow(), approval:null};
  progress.submissions[requirementId].richText ||= '';
  return progress.submissions[requirementId];
}

function configFor(userId, requirement) {
  const key = `${userId}:${requirement.id}`;
  const user=userById(userId), rules=(database.indicationRules||[]).filter(rule=>rule.active!==false&&rule.requirementId===requirement.id&&(
    rule.scope==='all'||(rule.scope==='club'&&rule.scopeId===user?.clubId)||(rule.scope==='role'&&rule.scopeId===user?.role)||(rule.scope==='user'&&rule.scopeId===userId)
  )).sort((a,b)=>({all:1,role:2,club:3,user:4}[a.scope]-({all:1,role:2,club:3,user:4}[b.scope])));
  return Object.assign({
    suggestedDate:'', mode:requirement.activity_mode || 'Individual', evidence:requirement.evidence_type_suggestion || 'Informe',
    photos:requirement.requires_photos_suggestion ? 'required' : 'optional', minWords:'', maxWords:'',
    instructions:requirement.tutor_instructions_default || '', evaluation:requirement.evaluation_suggestion || ''
  },...rules.map(rule=>rule.config||{}),database.configs[key]||{});
}

function countStatus(userId, status) { return Object.values(ensureProgress(userId).status).filter(value => value === status).length; }
function progressPercent(userId) { return Math.round(countStatus(userId, 'approved') * 100 / flatRequirements.length); }
function dailyPsalmNumber(date=new Date()){const anchor=Date.UTC(2026,7,26),current=Date.UTC(date.getFullYear(),date.getMonth(),date.getDate()),days=Math.round((current-anchor)/86400000);return ((18+days)%150+150)%150+1;}

function contextSelector() {
  const user = currentUser();
  if (['aspirant','pastor_aspirant'].includes(user.role)) return '';
  const aspirants = accessibleAspirants(user);
  if (!aspirants.length) return '<div class="notice warn">Todavía no hay Aspirantes asignados.</div>';
  return `<label class="field-group" style="min-width:280px;margin:0">Expediente del Aspirante<select class="field" onchange="selectAspirant(this.value)">${aspirants.map(aspirant => `<option value="${aspirant.id}" ${aspirant.id === activeAspirantId ? 'selected' : ''}>${esc(aspirant.fullName)}</option>`).join('')}</select></label>`;
}

function selectAspirant(id) { activeAspirantId = id; render(); }
function render() { go(currentPage); }

function sglHomeInspirationMarkup(){
  const item=window.SGL_INSPIRATION?.pick?.('home');
  if(!item)return '';
  const books=item.url?`<a class="sgl5065-home-inspiration-link" href="${esc(item.url)}" target="_blank" rel="noopener noreferrer">Libros de Elena G. de White ↗</a>`:'';
  return `<div class="sgl5065-home-inspiration" aria-label="Mensaje motivador"><blockquote>“${esc(item.text)}”</blockquote><div><strong>${esc(sgl5066DisplaySource(item))}</strong>${books}</div></div>`;
}
let sglSabbathHeroSequence=0;
const SGL_SABBATH_MESSAGES=[
  {text:'Acuérdate del día de reposo para santificarlo.',source:'Biblia · Éxodo 20:8'},
  {text:'El sábado fue hecho para el bien del ser humano.',source:'Biblia · Marcos 2:27, paráfrasis'},
  {text:'Llama al sábado delicia y honra el día santo del Señor.',source:'Biblia · Isaías 58:13, paráfrasis'},
  {text:'El sábado dirige la mente hacia el Creador y su obra redentora.',source:'Elena G. de White · El Deseado de Todas las Gentes, cap. 29 · idea resumida'},
  {text:'Las horas sagradas del sábado son una invitación a dejar lo común y acercarse a Dios.',source:'Elena G. de White · Testimonios para la Iglesia, t. 6 · idea resumida'},
  {text:'La naturaleza y la Palabra pueden conducir la atención de la familia hacia Dios durante el sábado.',source:'Elena G. de White · La Educación · idea resumida'}
];
function sglSabbathCoordinateKey(user=currentUser?.()){
  return `sgl_sabbath_coords_${String(user?.clubId||user?.id||'device')}`;
}
function sglSabbathCoordinatesFromMaps(url=''){
  let source='';try{source=decodeURIComponent(String(url||''));}catch{source=String(url||'');}
  const patterns=[/@(-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?)(?:,|$)/,/!3d(-?\d+(?:\.\d+)?)[^!]*!4d(-?\d+(?:\.\d+)?)/,/[?&](?:q|query|ll)=(-?\d+(?:\.\d+)?)[,%20 ]+(-?\d+(?:\.\d+)?)/i];
  for(const pattern of patterns){const match=source.match(pattern);if(!match)continue;const lat=Number(match[1]),lon=Number(match[2]);if(Number.isFinite(lat)&&Number.isFinite(lon)&&Math.abs(lat)<=90&&Math.abs(lon)<=180)return {lat,lon,source:'club_maps'};}
  return null;
}
function sglSabbathCoordinates(user=currentUser?.()){
  try{
    const club=clubFor?.(user)||null,lat=Number(club?.latitude??club?.lat),lon=Number(club?.longitude??club?.lng??club?.lon);
    if(Number.isFinite(lat)&&Number.isFinite(lon)&&Math.abs(lat)<=90&&Math.abs(lon)<=180)return {lat,lon,source:'club'};
    const maps=sglSabbathCoordinatesFromMaps(club?.mapsUrl||'');if(maps)return maps;
    const cached=JSON.parse(localStorage.getItem(sglSabbathCoordinateKey(user))||'null');
    if(cached&&Number.isFinite(Number(cached.lat))&&Number.isFinite(Number(cached.lon)))return {lat:Number(cached.lat),lon:Number(cached.lon),source:'device'};
  }catch{}
  return null;
}
async function sglSabbathPrimeCoordinates(){
  const user=currentUser?.();if(!user||sglSabbathCoordinates(user))return;
  if(!navigator.geolocation||!navigator.permissions?.query)return;
  try{
    const permission=await navigator.permissions.query({name:'geolocation'});if(permission.state!=='granted')return;
    navigator.geolocation.getCurrentPosition(position=>{
      const lat=Number(position?.coords?.latitude),lon=Number(position?.coords?.longitude);if(!Number.isFinite(lat)||!Number.isFinite(lon))return;
      try{localStorage.setItem(sglSabbathCoordinateKey(user),JSON.stringify({lat,lon,at:Date.now()}));}catch{}
    },()=>{},{enableHighAccuracy:false,maximumAge:86400000,timeout:5000});
  }catch{}
}
function sglSabbathSunset(date,coords){
  const fallback=new Date(date.getFullYear(),date.getMonth(),date.getDate(),18,0,0,0);
  if(!coords)return fallback;
  const rad=value=>value*Math.PI/180,deg=value=>value*180/Math.PI,normalize=value=>((value%360)+360)%360;
  // Calcula el número de día en UTC para que el cambio de horario estacional
  // no introduzca un desfase de un día en la fórmula solar.
  const day=Math.floor((Date.UTC(date.getFullYear(),date.getMonth(),date.getDate())-Date.UTC(date.getFullYear(),0,0))/86400000),lngHour=coords.lon/15,t=day+((18-lngHour)/24),m=(0.9856*t)-3.289;
  let l=normalize(m+(1.916*Math.sin(rad(m)))+(0.020*Math.sin(rad(2*m)))+282.634),ra=normalize(deg(Math.atan(0.91764*Math.tan(rad(l))))),lQuadrant=Math.floor(l/90)*90,raQuadrant=Math.floor(ra/90)*90;
  ra=(ra+(lQuadrant-raQuadrant))/15;
  const sinDec=0.39782*Math.sin(rad(l)),cosDec=Math.cos(Math.asin(sinDec)),cosH=(Math.cos(rad(90.833))-(sinDec*Math.sin(rad(coords.lat))))/(cosDec*Math.cos(rad(coords.lat)));
  if(!Number.isFinite(cosH)||cosH>1||cosH<-1)return fallback;
  const h=deg(Math.acos(cosH))/15,localMean=h+ra-(0.06571*t)-6.622,ut=((localMean-lngHour)%24+24)%24,base=Date.UTC(date.getFullYear(),date.getMonth(),date.getDate()),raw=base+ut*3600000,candidates=[raw-86400000,raw,raw+86400000];
  const sameLocalDay=candidates.find(ms=>{const value=new Date(ms);return value.getFullYear()===date.getFullYear()&&value.getMonth()===date.getMonth()&&value.getDate()===date.getDate();});
  if(sameLocalDay!==undefined)return new Date(sameLocalDay);
  const target=fallback.getTime();return new Date(candidates.sort((a,b)=>Math.abs(a-target)-Math.abs(b-target))[0]);
}
function sglSabbathContext(now=new Date(),user=currentUser?.()){
  const day=now.getDay(),coords=sglSabbathCoordinates(user);if(day!==5&&day!==6)return {active:false,coords};
  const sunset=sglSabbathSunset(now,coords);
  if(day===5&&now>=sunset)return {active:true,sunset,coords,friday:new Date(now.getFullYear(),now.getMonth(),now.getDate())};
  if(day===6&&now<sunset){const friday=new Date(now.getFullYear(),now.getMonth(),now.getDate()-1);return {active:true,sunset,coords,friday};}
  return {active:false,sunset,coords};
}
function sglSabbathMessage(context,user=currentUser?.()){
  const friday=context?.friday||new Date(),seed=Number(String(`${friday.getFullYear()}${friday.getMonth()+1}${friday.getDate()}`).replace(/\D/g,''))+(String(user?.id||'').length*7)+sglSabbathHeroSequence;
  return SGL_SABBATH_MESSAGES[Math.abs(seed)%SGL_SABBATH_MESSAGES.length];
}
function sgl5066WelcomeHeroBaseMarkup(greeting,firstName,subtitle='',now=new Date()){
  const dateText=now.toLocaleDateString('es',{weekday:'long',day:'numeric',month:'long',year:'numeric'}),monthText=now.toLocaleDateString('es',{month:'long'}).toUpperCase();
  return `<section class="welcome-hero sgl5066-welcome-hero"><div class="sgl5066-welcome-copy"><div class="eyebrow">${esc(dateText)}</div><h1>${esc(greeting)}, ${esc(firstName)}</h1>${subtitle?`<p>${esc(subtitle)}</p>`:''}</div>${sglHomeInspirationMarkup()}<div class="welcome-date"><b>${now.getDate()}</b><span>${esc(monthText)}</span></div></section>`;
}
function sgl5066WelcomeHeroMarkup(greeting,firstName,subtitle='',now=new Date()){
  const context=sglSabbathContext(now);if(!context.active)return sgl5066WelcomeHeroBaseMarkup(greeting,firstName,subtitle,now);
  const dateText=now.toLocaleDateString('es',{weekday:'long',day:'numeric',month:'long',year:'numeric'}),monthText=now.toLocaleDateString('es',{month:'long'}).toUpperCase(),message=sglSabbathMessage(context),heroId=`sgl_sabbath_hero_${++sglSabbathHeroSequence}`;
  setTimeout(()=>{const root=document.getElementById(heroId);if(root)root.outerHTML=sgl5066WelcomeHeroBaseMarkup(greeting,firstName,subtitle,new Date());},10000);
  return `<section id="${heroId}" class="welcome-hero sgl5066-welcome-hero sgl-sabbath-hero"><div class="sgl5066-welcome-copy"><div class="eyebrow">${esc(dateText)}</div><h1>Feliz sábado, ${esc(firstName)}</h1>${subtitle?`<p>${esc(subtitle)}</p>`:''}</div><div class="sgl5065-home-inspiration sgl-sabbath-inspiration" aria-label="Mensaje de sábado"><blockquote>${esc(message.text)}</blockquote><div><strong>${esc(message.source)}</strong></div></div><div class="welcome-date"><b>${now.getDate()}</b><span>${esc(monthText)}</span></div></section>`;
}
async function sglSabbathPoll(now=new Date()){
  const user=currentUser?.();if(!user)return false;const context=sglSabbathContext(now,user);if(!context.active)return false;
  const friday=context.friday,key=`sgl_sabbath_greeted_${String(user.id||'user')}_${friday.getFullYear()}-${String(friday.getMonth()+1).padStart(2,'0')}-${String(friday.getDate()).padStart(2,'0')}`;
  try{if(localStorage.getItem(key))return false;}catch{}
  // No consumir la entrega semanal antes de que exista al menos un canal
  // visible. Al iniciar sesión renderShell vuelve a ejecutar este sondeo.
  const canToast=!!document.getElementById('toast_root'),canNotify='Notification' in window&&Notification.permission==='granted';
  if(!canToast&&!canNotify)return false;
  const message='¡Feliz sábado! Que estas horas sean de descanso, comunión con Dios y renovación.';
  let delivered=false;
  try{if(canToast){toast(message);delivered=true;}}catch{}
  try{if(canNotify){new Notification('Feliz sábado',{body:message,icon:appLogo?.()||undefined});delivered=true;}}catch{}
  if(!delivered)return false;
  try{localStorage.setItem(key,isoNow());}catch{}
  return true;
}
setTimeout(()=>{sglSabbathPrimeCoordinates().catch(()=>{});sglSabbathPoll().catch(()=>{});},1200);
setInterval(()=>sglSabbathPoll().catch(()=>{}),60000);

function renderDashboard() {
  const user = currentUser();
  const aspirant = targetAspirant();
  setTop(['aspirant','pastor_aspirant'].includes(user.role) ? 'Panel de avance' : 'Panel de gestión', ROLE_LABELS[user.rawRole624||user.role]);
  if (!aspirant && !['aspirant','pastor_aspirant'].includes(user.role)) return setContent(`<div class="page-head"><div><div class="eyebrow">${ROLE_LABELS[user.rawRole624||user.role]}</div><h1 class="page-title">Bienvenido, ${esc(user.fullName)}</h1></div></div><div class="card empty">No hay Aspirantes disponibles. ${user.role === 'admin' ? 'Crea la primera cuenta desde Usuarios.' : 'El Administrador debe asignarte un Aspirante.'}</div>`);
  const approved = countStatus(aspirant.id, 'approved');
  const review = countStatus(aspirant.id, 'in_review');
  const developing = ['assigned','in_progress','needs_changes'].reduce((sum, status) => sum + countStatus(aspirant.id, status), 0);
  const pending = countStatus(aspirant.id, 'pending');
  const areaRows = MASTER.areas.map(area => {
    const completed = area.requirements.filter(req => ensureProgress(aspirant.id).status[req.id] === 'approved').length;
    const percent = Math.round(completed * 100 / area.requirements.length);
    return `<div class="area-row"><b>${esc(area.title)}</b><div class="bar"><i style="width:${percent}%"></i></div><span>${completed}/${area.requirements.length}</span></div>`;
  }).join('');
  const recent = Object.entries(ensureProgress(aspirant.id).submissions).sort((a,b) => String(b[1].updatedAt).localeCompare(String(a[1].updatedAt))).slice(0,6).map(([id, submission]) => {
    const requirement = flatRequirements.find(item => item.id === id);
    return `<div class="reading-item"><div><h4>${esc(id)} · ${esc(short(requirement?.official_text || '', 80))}</h4><p>${STATUS_LABELS[ensureProgress(aspirant.id).status[id]]} · ${formatDateTime(submission.updatedAt)}</p></div></div>`;
  }).join('');
  const now=new Date(),greeting=now.getHours()<12?'Buenos días':now.getHours()<19?'Buenas tardes':'Buenas noches',firstName=user.fullName.trim().split(/\s+/)[0]||user.fullName,dateText=now.toLocaleDateString('es',{weekday:'long',day:'numeric',month:'long',year:'numeric'}),monthText=now.toLocaleDateString('es',{month:'long'}).toUpperCase(),psalm=dailyPsalmNumber(now),conflictAlert=user.role==='admin'&&database.conflicts?.length?`<button class="notice warn dashboard-conflict-alert" onclick="go('settings')"><b>${database.conflicts.length} conflicto${database.conflicts.length===1?'':'s'} pendiente${database.conflicts.length===1?'':'s'}</b><span>Revisar en Configuración</span></button>`:'',tutorSummary=user.role==='tutor'?`<section class="card tutor-home-summary"><div class="card-head"><div><h3>Mis Aspirantes</h3><p class="subtitle" style="margin:0">Vista rápida de tus expedientes asignados.</p></div></div>${tutorAspirantMiniCards(user.id)}</section>`:'';
  setContent(`${sgl5066WelcomeHeroMarkup(greeting,firstName,'',now)}${conflictAlert}${tutorSummary}<div class="page-head"><div><div class="eyebrow">${['aspirant','pastor_aspirant'].includes(user.role) ? 'Mi progreso' : 'Seguimiento'}</div><h1 class="page-title">Expediente de ${esc(aspirant.fullName)}</h1></div>${contextSelector()}</div>${isBirthdayToday(aspirant)?`<div class="notice birthday">🎂 Hoy es el cumpleaños de <b>${esc(aspirant.fullName)}</b>.</div>`:''}<div class="grid4"><button class="card kpi kpi-link" onclick="openStatusRequirements('approved')"><span class="icon">✓</span><small>Revisados</small><strong>${approved}</strong><small>de 134 requisitos</small></button><button class="card kpi kpi-link" onclick="openStatusRequirements('in_review')"><span class="icon">⌛</span><small>En revisión</small><strong>${review}</strong><small>pendientes del Tutor</small></button><button class="card kpi kpi-link" onclick="openStatusRequirements('developing')"><span class="icon">✎</span><small>En desarrollo</small><strong>${developing}</strong><small>asignados o en corrección</small></button><button class="card kpi kpi-link" onclick="openStatusRequirements('pending')"><span class="icon">○</span><small>Pendientes</small><strong>${pending}</strong><small>por iniciar</small></button></div><div class="two"><div class="card"><div class="card-head"><h3>Avance por área</h3></div>${areaRows}</div><div class="card"><div class="card-head"><h3>Actividad reciente</h3></div>${recent || '<div class="empty">Todavía no hay actividad.</div>'}</div></div>`);
}

function openStatusRequirements(status) {
  requirementStatusFilter=status;
  go('requirements');
}

function accessibleClasses(user=currentUser()){
  return ClassPolicy.visibleClasses(user,database.classes||[]);
}

function classRequirements(item){return item.requirements||classCatalog[item.catalogKey||item.id]?.requirements||[];}
function classWork(item){const work=item.work||(item.work={});work.members=Array.isArray(work.members)?work.members:[];work.planning=work.planning&&typeof work.planning==='object'&&!Array.isArray(work.planning)?work.planning:{};if(Array.isArray(work.plans)){for(const old of work.plans){if(old.requirementId&&!work.planning[old.requirementId])work.planning[old.requirementId]={sessions:'',time:'',date:old.date||'',objective:old.objective||'',option:'',activities:old.title||'',materials:'',evaluation:'',cost:'',progressed:false};}delete work.plans;}work.attendance=Array.isArray(work.attendance)?work.attendance:[];work.tracking=work.tracking&&typeof work.tracking==='object'?work.tracking:{};for(const m of work.members){m.fullName||=m.name||'';m.phone||='';m.birthDate||='';m.address||='';m.fatherPhone||='';m.motherPhone||='';m.email||='';m.unit||='';m.photo||='';m.active=m.active!==false;}return work;}
function activeClassMembers(work){return (work?.members||[]).filter(m=>m.active!==false);}
function inactiveClassMembers(work){return (work?.members||[]).filter(m=>m.active===false);}

function renderClasses(){
  const user=currentUser(),canAssign=user?.role==='admin',items=accessibleClasses(user);
  setTop(['aspirant','pastor_aspirant'].includes(user.role)?'Mis clases':'Clases e instructores',ROLE_LABELS[user.rawRole624||user.role]);
  const cards=items.map(item=>{const assigned=item.assignedUserIds.map(userById).filter(Boolean),requirements=classRequirements(item).length;return `<article class="card class-card ${item.active===false?'class-disabled':''}" style="--class-accent:${classBranding(item).color};--class-accent-soft:${classBranding(item).soft}"><div class="class-number class-logo">${item.custom?'＋':`<img src="${esc(classBranding(item).logo)}" alt="${esc(item.regular)}">`}</div><div><div class="eyebrow">${item.custom?'Clase importada':'Clase regular y avanzada'} · ${esc(item.age||'—')} años</div><h3>${esc(item.regular)}</h3><p class="class-advanced">${esc(item.advanced||'Sin avanzada')}</p><div class="chips"><span class="chip blue">${assigned.length} instructor${assigned.length===1?'':'es'}</span><span class="chip">${requirements} requisitos</span><span class="chip ${item.active===false?'red':'green'}">${item.active===false?'Deshabilitada':'Habilitada'}</span></div></div><div class="class-actions">${canAssign?`<button class="btn ghost small" onclick="openClassConfig('${item.id}')">Asignar instructores</button>`:''}<button class="btn primary small" onclick="openClassWorkspace('${item.id}')">Abrir clase</button></div></article>`;}).join('');
  setContent(`<div class="page-head"><div><div class="eyebrow">Módulo interno</div><h1 class="page-title">${['aspirant','pastor_aspirant'].includes(user.role)?'Mis clases':'Clases'}</h1>${['aspirant','pastor_aspirant'].includes(user.role)?'':`<p class="subtitle">Solo el Administrador visualiza todas las clases y asigna Aspirantes como Instructores.</p>`}</div></div><div class="class-grid">${cards||'<div class="card empty">No hay clases disponibles.</div>'}</div>`);
}

function openClassConfig(id){
  const actor=currentUser(),item=(database.classes||[]).find(entry=>entry.id===id);if(!item||actor?.role!=='admin')return;
  const people=database.users.filter(person=>['aspirant','pastor_aspirant'].includes(person.role)&&(person.accountState||'active')==='active'&&person.active!==false);
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>${esc(item.regular)} + ${esc(item.advanced)}</h3><p>Solo el Administrador asigna Aspirantes. Dentro de esta clase se muestran como Instructores.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveClassConfig(event,'${item.id}')"><div class="modal-body"><label class="choice"><input id="class_active" type="checkbox" ${item.active!==false?'checked':''}> Clase habilitada</label><div class="section-label">Instructores asignados</div><div class="class-people">${people.map(person=>{const checked=item.assignedUserIds.includes(person.id),otherCount=ClassPolicy.assignmentCount(person.id,database.classes,item.id),total=otherCount+(checked?1:0),full=otherCount>=3&&!checked;return `<label class="choice ${full?'class-choice-disabled':''}"><input type="checkbox" name="class_instructor" value="${person.id}" ${checked?'checked':''} ${full?'disabled':''}> <span><b>${esc(person.fullName)}</b><small>${esc(clubFor(person)?.name||'Sin Club')} · ${total}/3 clases${full?' · límite alcanzado':''}</small></span></label>`;}).join('')||'<div class="notice warn">No hay Aspirantes activos disponibles para asignar.</div>'}</div><div class="notice" style="margin-top:14px">Un Aspirante puede ser Instructor en un máximo de 3 clases.</div></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar asignación</button></div></form></div>`);
}

async function saveClassConfig(event,id){
  event.preventDefault();const actor=currentUser(),item=(database.classes||[]).find(entry=>entry.id===id);if(!item||actor?.role!=='admin')return;
  const checked=[...document.querySelectorAll('[name=class_instructor]:checked')].map(input=>input.value),validation=ClassPolicy.validateAssignments(id,checked,database.classes,database.users,3);
  if(!validation.ok)return toast(validation.message||'No se pudo guardar la asignación.');
  item.assignedUserIds=validation.userIds;
  item.active=document.getElementById('class_active').checked;
  item.updatedAt=isoNow();await persist(true,{action:'Clase actualizada',entity:item.id,detail:`${item.name} · ${item.assignedUserIds.length} instructor(es)`});closeModal();renderClasses();toast('Clase e instructores actualizados.');
}

function classSections(){const memberLabel=currentUser().role==='admin'?'Conquis':'Mis Conquis';return [['dashboard','Inicio'],['members',memberLabel],['planning','Planificación'],['attendance','Asistencia'],['tracking','Seguimiento'],['stats','Estadísticas'],['calendar','Cronograma'],['instructors','Instructores'],['help','Cómo llenar']];}
function openClassWorkspace(id,section='dashboard'){if(!accessibleClasses().some(item=>item.id===id))return;activeClassId=id;activeClassSection=section;classAttendanceDate||=localDate();renderClassWorkspace();}
function classPlan(item,reqId){const w=classWork(item);return w.planning[reqId]||(w.planning[reqId]={sessions:'',time:'',date:'',objective:'',option:'',activities:'',materials:'',evaluation:'',cost:'',progressed:false});}
function classPlanProgress(item,reqs){const done=reqs.filter(r=>classPlan(item,r.id).progressed).length;return{done,total:reqs.length,pct:reqs.length?Math.round(done*100/reqs.length):0};}
function classTrackComplete(work,memberId,reqId){const v=work.tracking[`${memberId}:${reqId}`];return v===true||v==='completed';}
function classAttendanceEntries(record,item){if(record.entries)return record.entries;const entries={};for(const id of record.presentIds||[])entries[`m:${id}`]='Presente';return entries;}
function classAttendanceRate(item){const w=classWork(item);let total=0,present=0;for(const r of w.attendance){for(const status of Object.values(classAttendanceEntries(r,item))){if(!status)continue;total++;if(status==='Presente'||status==='Tarde')present++;}}return total?Math.round(present*100/total):0;}
function classReqLabel(req){return req.kind==='advanced'?'Avanzada':'Regular';}
function classAreaAbbreviation(area,req={}){
  const cleaned=String(area||'').trim().replace(/^CLASE\s+AVANZADA\s*[·:-]?\s*/i,''),roman=(cleaned.match(/^([IVXLCDM]+)\.(?:\s|$)/i)||[])[1]?.toUpperCase()||'',name=cleaned.replace(/^([IVXLCDM]+)\.(?:\s|$)/i,'');
  const stop=new Set(['A','AL','DE','DEL','LA','LAS','EL','LOS','Y','E','EN','CON','PARA','POR','UN','UNA','UNO','UNOS','UNAS']);
  const initials=name.normalize('NFD').replace(/[\u0300-\u036f]/g,'').split(/[^A-Za-z0-9]+/).filter(Boolean).filter(word=>!stop.has(word.toUpperCase())).map(word=>word[0].toUpperCase()).join('');
  return roman?`${roman}-${initials||roman}`:(initials||'R');
}
function classRequirementShort(req,max=58){let text=String(req?.requirement||req?.official_text||'').replace(/\s+/g,' ').replace(/^\s*\d+[.)-]?\s*/,'').trim();const sentence=text.match(/^(.+?[.!?])(?:\s|$)/)?.[1];if(sentence&&sentence.length<=max)text=sentence.replace(/[.!?]+$/,'');if(text.length>max)text=text.slice(0,max-1).replace(/\s+\S*$/,'').trim()+'…';return text;}
function classRequirementVerticalLines(req,max=46){const text=classRequirementShort(req,max),ellipsis=text.endsWith('…'),base=text.replace(/…$/,'').trim(),words=base.split(/\s+/).filter(Boolean);if(words.length<3||base.length<=22)return[text];const target=Math.ceil(base.length/2);let first='',second='';for(const word of words){const candidate=(first?first+' ':'')+word;if(!second&&candidate.length<=target)first=candidate;else second+=(second?' ':'')+word;}if(!second)return[text];return[first,second+(ellipsis?'…':'')];}
function filterClassRequirements(reqs,filter='all'){return filter==='regular'?reqs.filter(r=>r.kind==='regular'):filter==='advanced'?reqs.filter(r=>r.kind==='advanced'):reqs;}
function classMonthlyActivities(item,date=new Date()){
  const work=classWork(item),reqs=classRequirements(item),byId=new Map(reqs.map(req=>[req.id,req]));
  const now=date instanceof Date?date:new Date(date),today=(()=>{const copy=new Date(now.getTime()-now.getTimezoneOffset()*60000);return copy.toISOString().slice(0,10);})(),month=today.slice(0,7);
  return Object.entries(work.planning).map(([requirementId,plan])=>{const req=byId.get(requirementId);if(!req||!plan?.date||!String(plan.date).startsWith(month))return null;const status=plan.date===today?'today':plan.date<today?'overdue':'upcoming';return{requirementId,date:plan.date,status,requirement:req,plan};}).filter(Boolean).sort((a,b)=>String(a.date).localeCompare(String(b.date)));
}
function classInstructorBirthdayNotice(item,assigned=[]){
  const candidates=(assigned&&assigned.length?assigned:(item?.assignedUserIds||[]).map(userById).filter(Boolean)).filter(person=>person?.birthDate);
  const instructors=window.SGLViewPolicy562?.birthdays(candidates,currentUser(),database,'mc')||[];
  if(!instructors.length)return '';
  const today=instructors.filter(isBirthdayToday);
  const list=today.length?today:instructors.slice().sort((a,b)=>String(a.birthDate||'').slice(5).localeCompare(String(b.birthDate||'').slice(5))).slice(0,3);
  const label=today.length?'Hoy cumple':'Cumpleaños de instructores';
  const cls=today.length?'birthday':'offline';
  return `<div class="notice ${cls} class-birthday-notice"><b>🎂 ${esc(label)}:</b> ${list.map(person=>`<span>${esc(person.fullName)}${person.birthDate?` · ${String(person.birthDate).slice(5).split('-').reverse().join('/')}`:''}</span>`).join(' · ')}</div>`;
}

function renderClassWorkspace(){
  const item=accessibleClasses().find(entry=>entry.id===activeClassId);if(!item)return renderClasses();const brand=classBranding(item),work=classWork(item),activeMembers=activeClassMembers(work),reqs=classRequirements(item),regular=reqs.filter(r=>r.kind==='regular'),advanced=reqs.filter(r=>r.kind==='advanced'),rp=classPlanProgress(item,regular),ap=classPlanProgress(item,advanced),allp=classPlanProgress(item,reqs),assigned=item.assignedUserIds.map(userById).filter(Boolean),attendance=classAttendanceRate(item);
  setTop(activeClassSection==='tracking'?'Seguimiento':`${item.regular} + ${item.advanced||''}`,activeClassSection==='tracking'?'':'Clase integrada');
  const tabs=`<div class="class-tabs">${classSections().map(([id,label])=>`<button class="${activeClassSection===id?'active':''}" onclick="openClassWorkspace('${item.id}','${id}')">${label}</button>`).join('')}</div>`;
  let body='';
  if(activeClassSection==='dashboard'){
    const areas=[...new Set(reqs.map(r=>r.area))];
    const monthly=classMonthlyActivities(item),monthlyRows=monthly.slice(0,8).map(activity=>`<div class="class-month-activity ${activity.status}"><span class="class-month-date">${formatDate(activity.date)}</span><div><b>${esc(classAreaAbbreviation(activity.requirement.area,activity.requirement))} · ${esc(classRequirementShort(activity.requirement,62))}</b><small>${activity.status==='today'?'Hoy':activity.status==='overdue'?'Fecha pasada':'Próxima actividad'}${activity.plan.sessions?` · Sesión ${esc(activity.plan.sessions)}`:''}</small></div></div>`).join('');
    body=`<div class="page-head"><div><h2>${esc(item.regular)} · ${esc(item.age||'—')} años</h2><p class="subtitle">La clase regular y ${esc(item.advanced||'su avanzada')} se trabajan como una sola asignación para el Instructor.</p></div><div class="actions"><button class="btn primary" onclick="openClassWorkspace('${item.id}','planning')">Abrir planificación</button><button class="btn ghost" onclick="openClassWorkspace('${item.id}','members')">Gestionar miembros</button></div></div>${classInstructorBirthdayNotice(item,assigned)}<div class="grid4"><div class="card kpi"><small>Conquistadores</small><strong>${activeMembers.length}</strong><small>miembros activos</small></div><div class="card kpi"><small>Clase Regular</small><strong>${rp.pct}%</strong><small>${rp.done}/${rp.total} requisitos</small></div><div class="card kpi"><small>Clase avanzada</small><strong>${ap.pct}%</strong><small>${ap.done}/${ap.total} requisitos</small></div><div class="card kpi"><small>Asistencia</small><strong>${attendance}%</strong><small>Presente + Tarde</small></div></div><div class="two"><div class="card"><h3>Avance por área</h3>${areas.map(area=>{const p=classPlanProgress(item,reqs.filter(r=>r.area===area));return `<div class="class-progress-row"><span title="${esc(area)}">${esc(short(area,35))}</span><div class="bar"><i style="width:${p.pct}%"></i></div><b>${p.pct}%</b></div>`;}).join('')}</div><div class="card"><div class="card-head"><div><h3>Actividades del mes</h3><p class="subtitle" style="margin:0">Se generan automáticamente desde las fechas de Planificación.</p></div><button class="btn ghost small" onclick="openClassWorkspace('${item.id}','calendar')">Ver cronograma</button></div>${monthlyRows||'<div class="empty">No hay actividades planificadas para este mes.</div>'}</div></div>`;
  }
  if(activeClassSection==='members'){
    const q=normalize(classMemberSearch),activeList=activeClassMembers(work),inactiveList=inactiveClassMembers(work),source=classMemberFilter==='inactive'?inactiveList:activeList,members=source.filter(m=>!q||normalize(`${m.fullName} ${m.unit} ${m.email}`).includes(q)),admin=currentUser().role==='admin';
    body=`<div class="page-head"><div><h2>${currentUser().role==='admin'?'Conquis':'Mis Conquis'}</h2><p class="subtitle">Galería de ${esc(item.regular)} · abre una tarjeta para consultar o editar sus datos.</p></div><div class="actions">${admin?`<button class="btn red" onclick="resetClassProgress('${item.id}')">Reiniciar clase</button>`:''}<button class="btn primary" onclick="openClassMember()">+ Añadir miembro</button></div></div><div class="card class-toolbar class-member-toolbar"><div class="class-member-filters"><button class="btn small ${classMemberFilter==='active'?'primary':'ghost'}" onclick="classMemberFilter='active';renderClassWorkspace()">Activos <span>${activeList.length}</span></button><button class="btn small ${classMemberFilter==='inactive'?'primary':'ghost'}" onclick="classMemberFilter='inactive';renderClassWorkspace()">Inhabilitados <span>${inactiveList.length}</span></button></div><input class="field" placeholder="Buscar por nombre, unidad o correo" value="${esc(classMemberSearch)}" oninput="classMemberSearch=this.value;renderClassWorkspace()"></div><div class="class-member-gallery">${members.map(m=>`<article class="card class-member-tile ${m.active===false?'inactive':''}"><button class="class-member-face" onclick="openClassMember('${m.id}')" title="Abrir ${esc(m.fullName)}">${m.photo?`<img src="${esc(m.photo)}" alt="${esc(m.fullName)}">`:`<span>${esc(initials(m.fullName))}</span>`}</button><div class="class-member-caption"><h3 title="${esc(m.fullName)}">${esc(m.fullName)}</h3><p>${esc(m.unit||'Sin unidad')}</p></div><div class="class-member-tile-actions"><button class="btn ghost small" onclick="openClassMember('${m.id}')">Editar</button>${m.active===false?`<button class="btn primary small" onclick="setClassMemberActive('${item.id}','${m.id}',true)">Habilitar</button>`:`<button class="btn ghost small" onclick="setClassMemberActive('${item.id}','${m.id}',false)">Inhabilitar</button>`}${admin?`<button class="btn red small" onclick="deleteClassMember('${m.id}')">Eliminar de clase</button>`:''}</div></article>`).join('')||`<div class="card empty">${classMemberFilter==='inactive'?'No hay miembros inhabilitados.':'Aún no hay miembros activos.'}</div>`}</div>`;
  }
  if(activeClassSection==='planning'){
    const visible=filterClassRequirements(reqs,classPlanFilter),rowFor=r=>{const p=classPlan(item,r.id);return `<tr><td><input type="checkbox" ${p.progressed?'checked':''} onchange="saveClassPlanField('${item.id}','${r.id}','progressed',this.checked)"></td><td>${esc(r.area)}</td><td class="class-req-cell">${esc(r.requirement)}<div class="chips"><span class="chip ${r.kind==='advanced'?'blue':''}">${classReqLabel(r)}</span><span class="chip">Pág. ${esc(r.page??'—')}</span></div></td>${[['sessions','text'],['time','text'],['date','date']].map(([f,t])=>`<td><input class="field compact class-table-input" type="${t}" value="${esc(p[f])}" oninput="saveClassPlanField('${item.id}','${r.id}','${f}',this.value)"></td>`).join('')}${['objective','option','activities','materials','evaluation'].map(f=>`<td><textarea class="field compact class-table-area" oninput="saveClassPlanField('${item.id}','${r.id}','${f}',this.value)">${esc(p[f])}</textarea></td>`).join('')}<td><input class="field compact class-table-input" type="number" min="0" step="0.01" value="${esc(p.cost)}" oninput="saveClassPlanField('${item.id}','${r.id}','cost',this.value)"></td></tr>`;},regularRows=visible.filter(r=>r.kind==='regular'),advancedRows=visible.filter(r=>r.kind==='advanced'),rows=`${classPlanFilter==='all'&&regularRows.length?`<tr class="class-group-row"><td colspan="12">CLASE REGULAR · ${esc(item.regular)}</td></tr>`:''}${regularRows.map(rowFor).join('')}${classPlanFilter==='all'&&advancedRows.length?`<tr class="class-group-row advanced"><td colspan="12">CLASE AVANZADA · ${esc(item.advanced||'')}</td></tr>`:''}${advancedRows.map(rowFor).join('')}`;
    body=`<div class="page-head"><div><h2>Plan de clase ${esc(item.regular)}</h2><p class="subtitle">Regular + ${esc(item.advanced||'avanzada')} · cambios guardados por campo.</p></div><div class="actions"><button class="btn small ${classPlanFilter==='all'?'primary':'ghost'}" onclick="classPlanFilter='all';renderClassWorkspace()">Ambas</button><button class="btn small ${classPlanFilter==='regular'?'primary':'ghost'}" onclick="classPlanFilter='regular';renderClassWorkspace()">Regular</button><button class="btn small ${classPlanFilter==='advanced'?'primary':'ghost'}" onclick="classPlanFilter='advanced';renderClassWorkspace()">Avanzada</button></div></div><div class="chips class-plan-progress"><span class="chip green">Total ${allp.pct}%</span><span class="chip">Regular ${rp.pct}%</span><span class="chip blue">Avanzada ${ap.pct}%</span></div><section class="card class-table-card"><div class="class-table-wrap"><table class="table class-plan-table"><thead><tr><th>Avanzó</th><th>Área</th><th>Requisito</th><th>N° sesión</th><th>T</th><th>Fecha</th><th>Objetivo</th><th>Opción / especialidad</th><th>Metodología / actividades</th><th>Materiales</th><th>Evaluación</th><th>Costo</th></tr></thead><tbody>${rows}</tbody></table></div></section>`;
  }
  if(activeClassSection==='attendance'){
    const date=classAttendanceDate||localDate(),record=work.attendance.find(r=>r.date===date),entries=record?classAttendanceEntries(record,item):{},people=[...activeMembers.map(m=>({key:`m:${m.id}`,name:m.fullName,type:'Conquistador',detail:m.unit||'Sin unidad',photo:m.photo||''})),...assigned.map(u=>({key:`i:${u.id}`,name:u.fullName,type:'Instructor',detail:'—',photo:u.photo||''}))],history=work.attendance.slice().sort((a,b)=>String(b.date).localeCompare(String(a.date)));
    body=`<div class="page-head"><div><h2>Asistencia de ${esc(item.regular)}</h2><p class="subtitle">Cada integrante ocupa una sola fila para marcar la reunión rápidamente.</p></div></div><div class="class-att-grid"><section class="card"><div class="field-row"><label class="field-group">Fecha<input class="field" type="date" value="${esc(date)}" onchange="classAttendanceDate=this.value;renderClassWorkspace()"></label><label class="field-group">Nota de la reunión<input id="class_att_note" class="field" value="${esc(record?.note||'')}"></label></div><div class="class-table-wrap"><table class="table class-attendance-table"><thead><tr><th>Nombre</th><th>Unidad / detalle</th><th>Tipo</th><th>Estado</th></tr></thead><tbody>${people.map(p=>`<tr><td><div class="class-att-person"><span class="class-att-avatar">${p.photo?`<img src="${esc(p.photo)}" alt="">`:esc(initials(p.name))}</span><b>${esc(p.name)}</b></div></td><td>${esc(p.detail)}</td><td>${p.type}</td><td><select class="field compact class-att-status" data-person="${esc(p.key)}">${['','Presente','Ausente','Tarde','Justificado'].map(st=>`<option ${entries[p.key]===st?'selected':''}>${st||'Sin marcar'}</option>`).join('')}</select></td></tr>`).join('')||'<tr><td colspan="4">No hay integrantes.</td></tr>'}</tbody></table></div><button class="btn primary" style="margin-top:12px" onclick="saveClassAttendanceRecord('${item.id}')" ${people.length?'':'disabled'}>Guardar asistencia</button></section><section class="card"><h3>Historial</h3>${history.map(r=>{const vals=Object.values(classAttendanceEntries(r,item)).filter(Boolean),pres=vals.filter(v=>v==='Presente'||v==='Tarde').length,pct=vals.length?Math.round(pres*100/vals.length):0;return `<div class="reading-item"><div><b>${formatDate(r.date)}</b><p>${esc(r.note||'Sin nota')}</p></div><button class="btn ghost small" onclick="classAttendanceDate='${r.date}';renderClassWorkspace()">${pct}% · ${pres}/${vals.length}</button></div>`;}).join('')||'<div class="empty">Sin historial.</div>'}</section></div>`;
  }
  if(activeClassSection==='tracking'){
    const visible=filterClassRequirements(reqs,classTrackFilter),regularVisible=visible.filter(r=>r.kind==='regular'),advancedVisible=visible.filter(r=>r.kind==='advanced');
    const areaGroups=list=>{const groups=[];for(const req of list){const last=groups.at(-1);if(last&&last.area===req.area)last.count+=1;else groups.push({area:req.area,count:1,req});}return groups;};
    const kindHeaders=`${regularVisible.length?`<th class="class-track-kind regular" colspan="${regularVisible.length}">REGULAR · ${esc(item.regular)}</th>`:''}${advancedVisible.length?`<th class="class-track-kind advanced" colspan="${advancedVisible.length}">AVANZADA · ${esc(item.advanced||'')}</th>`:''}`;
    const areaHeaders=[...areaGroups(regularVisible),...areaGroups(advancedVisible)].map(group=>`<th class="class-track-area" colspan="${group.count}" title="${esc(group.area)}">${esc(group.area)}</th>`).join('');
    const reqHeaders=visible.map(r=>`<th class="class-track-req-head" title="${esc(r.requirement)}"><span class="class-track-vertical">${classRequirementVerticalLines(r,46).map(line=>`<span>${esc(line)}</span>`).join('')}</span></th>`).join('');
    const progressFor=m=>{const done=visible.filter(r=>classTrackComplete(work,m.id,r.id)).length;return visible.length?Math.round(done*100/visible.length):0;};
    const rows=activeMembers.map(m=>`<tr><td class="class-track-member-cell"><div class="class-track-person"><span class="class-track-avatar">${m.photo?`<img src="${esc(m.photo)}" alt="">`:esc(initials(m.fullName))}</span><span><b>${esc(m.fullName)}</b><small>${esc(m.unit||'Sin unidad')}</small></span></div></td>${visible.map(r=>`<td class="class-track-mark" title="${esc(r.requirement)}"><input type="checkbox" aria-label="${esc(m.fullName)} · ${esc(classRequirementShort(r,34))}" ${classTrackComplete(work,m.id,r.id)?'checked':''} onchange="saveClassTrackingCheck('${item.id}','${m.id}','${r.id}',this.checked)"></td>`).join('')}<td class="class-track-progress">${progressFor(m)}%</td></tr>`).join('');
    body=`<div class="class-track-toolbar"><div class="actions"><button class="btn small ${classTrackFilter==='all'?'primary':'ghost'}" onclick="classTrackFilter='all';renderClassWorkspace()">Ambas</button><button class="btn small ${classTrackFilter==='regular'?'primary':'ghost'}" onclick="classTrackFilter='regular';renderClassWorkspace()">Regular</button><button class="btn small ${classTrackFilter==='advanced'?'primary':'ghost'}" onclick="classTrackFilter='advanced';renderClassWorkspace()">Avanzada</button></div></div>${activeMembers.length&&visible.length?`<section class="card class-track-card"><div class="class-table-wrap class-track-scroll"><table class="class-track-matrix"><thead><tr><th class="class-track-corner" rowspan="3">CONQUISTADORES</th>${kindHeaders}<th class="class-track-progress-head" rowspan="3">AVANCE</th></tr><tr>${areaHeaders}</tr><tr>${reqHeaders}</tr></thead><tbody>${rows}</tbody></table></div><div class="class-track-legend"><span>✓ Marca cuando el requisito esté completado.</span><span>Pasa el cursor sobre la columna para ver el requisito completo.</span></div></section>`:'<div class="card empty">Añade miembros y verifica que existan requisitos en este filtro para habilitar el seguimiento.</div>'}`;
  }
  if(activeClassSection==='stats'){
    const areas=[...new Set(reqs.map(r=>r.area))],row=rs=>{const p=classPlanProgress(item,rs),cost=rs.reduce((sum,r)=>sum+(Number(classPlan(item,r.id).cost)||0),0),sessions=rs.filter(r=>String(classPlan(item,r.id).sessions).trim()).length;let mp=0;if(activeMembers.length&&rs.length){let marks=0;for(const m of activeMembers)marks+=rs.filter(r=>classTrackComplete(work,m.id,r.id)).length;mp=Math.round(marks*100/(activeMembers.length*rs.length));}return{...p,cost,sessions,mp};},overall=row(reqs);
    body=`<div class="page-head"><div><h2>Estadísticas de avance</h2><p class="subtitle">Resumen del plan, seguimiento individual y asistencia.</p></div></div><div class="grid4"><div class="card kpi"><small>Requisitos</small><strong>${reqs.length}</strong></div><div class="card kpi"><small>Completados</small><strong>${overall.done}</strong><small>${reqs.length-overall.done} pendientes</small></div><div class="card kpi"><small>Avance plan</small><strong>${overall.pct}%</strong></div><div class="card kpi"><small>Asistencia</small><strong>${attendance}%</strong></div></div><div class="class-stats-grid"><section class="card class-table-card"><h3>Resumen por área</h3><div class="class-table-wrap"><table class="table"><thead><tr><th>Área</th><th>Total</th><th>Completados</th><th>Pendientes</th><th>% plan</th><th>Prom. miembros</th><th>Costo</th><th>Con sesión</th></tr></thead><tbody>${areas.map(a=>{const x=row(reqs.filter(r=>r.area===a));return `<tr><td>${esc(a)}</td><td>${x.total}</td><td>${x.done}</td><td>${x.total-x.done}</td><td>${x.pct}%</td><td>${x.mp}%</td><td>${x.cost.toFixed(2)}</td><td>${x.sessions}</td></tr>`;}).join('')}</tbody></table></div></section><section class="card class-table-card"><h3>Avance por Conquistador</h3><div class="class-table-wrap"><table class="table"><thead><tr><th>Conquistador</th><th>Regular</th><th>Avanzada</th><th>Total</th></tr></thead><tbody>${activeMembers.map(m=>{const calc=rs=>rs.length?Math.round(rs.filter(r=>classTrackComplete(work,m.id,r.id)).length*100/rs.length):0;return `<tr><td><b>${esc(m.fullName)}</b><br><small>${esc(m.unit||'Sin unidad')}</small></td><td>${calc(regular)}%</td><td>${calc(advanced)}%</td><td><b>${calc(reqs)}%</b></td></tr>`;}).join('')||'<tr><td colspan="4">Sin miembros.</td></tr>'}</tbody></table></div></section></div>`;
  }
  if(activeClassSection==='calendar'){
    const schedule=database.settings.clubSchedule,monthly=classMonthlyActivities(item),admin=currentUser().role==='admin';
    body=`<div class="page-head"><div><h2>Cronograma del Club</h2><p class="subtitle">Documento oficial del Club y actividades mensuales derivadas de la planificación.</p></div>${admin?`<label class="btn primary">${schedule?'Reemplazar cronograma':'Subir cronograma'}<input hidden type="file" accept=".pdf,image/*,.doc,.docx,.xls,.xlsx,.ppt,.pptx" onchange="saveClubSchedule(this.files[0]);this.value=''"></label>`:''}</div><div class="two"><section class="card"><div class="card-head"><div><h3>Archivo oficial</h3><p class="subtitle" style="margin:0">${schedule?`${esc(schedule.fileName)} · actualizado ${formatDateTime(schedule.updatedAt)}`:'El Administrador todavía no ha cargado un cronograma.'}</p></div>${schedule?`<div class="actions"><button class="btn primary small" onclick="openClubSchedule()">Visualizar</button>${admin?'<button class="btn red small" onclick="removeClubSchedule()">Retirar</button>':''}</div>`:''}</div>${schedule?'<div class="class-calendar-file">▣<span>Disponible para todos los Instructores asignados</span></div>':'<div class="empty">Sin archivo de cronograma.</div>'}</section><section class="card"><h3>Actividades del mes</h3>${monthly.map(activity=>`<div class="class-month-activity ${activity.status}"><span class="class-month-date">${formatDate(activity.date)}</span><div><b>${esc(classAreaAbbreviation(activity.requirement.area,activity.requirement))} · ${esc(classRequirementShort(activity.requirement,62))}</b><small>${activity.status==='today'?'Hoy':activity.status==='overdue'?'Fecha pasada':'Próxima actividad'}${activity.plan.sessions?` · Sesión ${esc(activity.plan.sessions)}`:''}</small></div></div>`).join('')||'<div class="empty">No hay actividades planificadas este mes.</div>'}</section></div>`;
  }
  if(activeClassSection==='instructors')body=`<div class="page-head"><div><h2>Instructores de ${esc(item.regular)}</h2><p class="subtitle">Son Aspirantes existentes asignados por el Administrador.</p></div>${currentUser().role==='admin'?`<button class="btn primary" onclick="openClassConfig('${item.id}')">Asignar instructores</button>`:'<span class="chip">Solo lectura</span>'}</div><div class="grid3">${assigned.map(u=>`<div class="card user-row">${avatarMarkup(u)}<div><h3>${esc(u.fullName)}</h3><p>${esc(u.email||'')}</p><span class="chip">${ClassPolicy.assignmentCount(u.id,database.classes)}/3 clases</span></div></div>`).join('')||'<div class="card empty">Sin Instructores asignados.</div>'}</div>`;
  if(activeClassSection==='help')body=`<div class="page-head"><div><h2>Cómo llenar la planificación</h2><p class="subtitle">Guía práctica para el Instructor de Clase.</p></div></div><div class="notice"><b>Material de apoyo.</b> Usa esta guía para mantener una planificación clara y uniforme.</div><div class="class-help-grid">${[['N° sesión','Registra en qué sesión se trabajó el requisito.'],['T · Tiempo','Usa una convención breve: 1D, 1H, 35M.'],['Fecha','Fecha en que el requisito fue culminado.'],['Requisito y página','Ya vienen cargados desde el catálogo oficial.'],['Objetivo','Resultado concreto que espera lograr el Instructor.'],['Opción / especialidad','Inciso, opción o especialidad elegida.'],['Metodología / actividades','Dinámica, exposición, salida, proyecto, demostración o lectura.'],['Materiales','Elementos necesarios para realizar la actividad.'],['Evaluación','Cómo se comprobará el aprendizaje o cumplimiento.'],['Costo','Costo aproximado o real de la actividad.'],['Avanzó: Sí / No','Marca Sí cuando la clase haya trabajado ese requisito.'],['Seguimiento individual','Marca requisito por requisito para cada Conquistador; es independiente de Avanzó.']].map(([t,x])=>`<section class="card"><h3>${t}</h3><p>${x}</p></section>`).join('')}</div>`;
  setContent(`<section class="class-workspace${activeClassSection==='tracking'?' tracking-compact':''}" style="--class-accent:${brand.color};--class-accent-2:${brand.color2};--class-accent-soft:${brand.soft};--accent:${brand.color};--accent-2:${brand.color2};--accent-soft:${brand.soft}"><div class="class-work-head"><button class="btn ghost" onclick="renderClasses()">← Volver a Clases</button><img class="class-work-logo" src="${esc(brand.logo)}" alt="${esc(item.regular)}"><div><div class="eyebrow">Clase regular y avanzada · ${esc(item.age||'—')} años</div><h1>${esc(item.regular)} <span>+ ${esc(item.advanced||'Sin avanzada')}</span></h1></div></div>${tabs}<div class="class-work-body">${body}</div></section>`);
}
async function saveClassPlanField(id,reqId,field,val){const item=accessibleClasses().find(x=>x.id===id);if(!item)return;classPlan(item,reqId)[field]=val;await persist(true,{action:'Planificación de clase actualizada',entity:id,detail:`${reqId}:${field}`});if(field==='progressed')renderClassWorkspace();}
function openClassMember(memberId=''){
  const item=accessibleClasses().find(x=>x.id===activeClassId);if(!item)return;
  const member=memberId?classWork(item).members.find(m=>m.id===memberId):null;
  const photoPreview=member?avatarMarkup(member):'📷';
  const photoActions=member?`<div class="field-group full"><span>Fotografía actual</span><div class="user-row"><div class="avatar" id="class_member_photo_preview">${photoPreview}</div><div class="actions"><button type="button" class="btn ghost small" onclick="openClassMemberPhotoStudio('${item.id}','${member.id}')">Estudio de foto</button><label class="btn ghost small">Reemplazar rápido<input id="class_member_photo" type="file" accept="image/*" onchange="previewClassMemberPhoto(this)" hidden></label><small id="class_member_photo_name" class="muted">Galería o archivos del teléfono</small></div></div></div>`:`<div class="field-group full"><span>Fotografía</span><input id="class_member_photo" type="file" accept="image/*" class="field" onchange="previewClassMemberPhoto(this)"><div class="user-row" style="margin-top:8px"><div class="avatar" id="class_member_photo_preview">${photoPreview}</div><small id="class_member_photo_name" class="muted">Selecciona una foto desde la galería o archivos.</small></div></div>`;
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>${member?'Editar':'Añadir'} Conquistador</h3><p>${esc(item.regular)}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveClassMember(event,'${memberId}')"><div class="modal-body"><div class="field-row"><label class="field-group full">Nombres y Apellidos Completos<input id="class_member_name" class="field" required autocomplete="off" list="class_name_history" value="${esc(member?.fullName||'')}"></label><datalist id="class_name_history">${nameAutocompleteOptions()}</datalist><div class="field-group full"><button type="button" class="btn link small" onclick="clearNameAutocompleteHistory()">Limpiar nombres guardados</button></div><label class="field-group">Celular<input id="class_member_phone" class="field" value="${esc(member?.phone||'')}"></label><label class="field-group">Fecha de nacimiento<input id="class_member_birth" type="date" class="field" value="${esc(member?.birthDate||'')}"></label><label class="field-group full">Dirección<input id="class_member_address" class="field" value="${esc(member?.address||'')}"></label><label class="field-group">Teléfono papá<input id="class_member_father" class="field" value="${esc(member?.fatherPhone||'')}"></label><label class="field-group">Teléfono mamá<input id="class_member_mother" class="field" value="${esc(member?.motherPhone||'')}"></label><label class="field-group">Correo<input id="class_member_email" type="email" class="field" value="${esc(member?.email||'')}"></label><label class="field-group">Unidad<input id="class_member_unit" class="field" value="${esc(member?.unit||'')}"></label>${photoActions}</div></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar</button></div></form></div>`);
}
function previewClassMemberPhoto(input){
  const file=input?.files?.[0],preview=document.getElementById('class_member_photo_preview'),name=document.getElementById('class_member_photo_name');
  if(!file)return;
  if(file.type&&!file.type.startsWith('image/')){input.value='';return toast('Selecciona una fotografía válida.');}
  if(name)name.textContent=file.name||'Foto seleccionada';
  if(!preview)return;
  const url=URL.createObjectURL(file);
  preview.innerHTML=`<img src="${url}" alt="Vista previa">`;
  const image=preview.querySelector('img');if(image)image.onload=()=>URL.revokeObjectURL(url);
}
async function saveClassMember(event,memberId=''){
  event.preventDefault();const item=accessibleClasses().find(x=>x.id===activeClassId);if(!item)return;
  const w=classWork(item),m=memberId?w.members.find(x=>x.id===memberId):{id:newId('member'),photo:'',active:true};
  m.fullName=value('class_member_name').trim();m.phone=value('class_member_phone').trim();m.birthDate=value('class_member_birth');m.address=value('class_member_address').trim();m.fatherPhone=value('class_member_father').trim();m.motherPhone=value('class_member_mother').trim();m.email=value('class_member_email').trim();m.unit=value('class_member_unit').trim();
  const photo=document.getElementById('class_member_photo')?.files?.[0];
  if(photo){try{m.photo=await makeSquareProfilePhoto(photo);}catch(error){await logClientError('Foto de Conquistador',error);return toast('No se pudo procesar esa foto. Prueba con una imagen JPG, PNG o WEBP.');}}
  if(!memberId)w.members.push(m);rememberAutocompleteName(m.fullName);
  await persist(true,{action:memberId?'Miembro de clase actualizado':'Miembro de clase agregado',entity:item.id,detail:m.fullName});closeModal();renderClassWorkspace();
}
async function setClassMemberActive(classId,memberId,active){const item=accessibleClasses().find(x=>x.id===classId);if(!item)return;const member=classWork(item).members.find(x=>x.id===memberId);if(!member)return;member.active=!!active;await persist(true,{action:active?'Miembro de clase habilitado':'Miembro de clase inhabilitado',entity:item.id,detail:member.fullName});renderClassWorkspace();}
async function deleteClassMember(memberId){if(currentUser()?.role!=='admin')return toast('Solo el Administrador puede eliminar definitivamente un miembro.');const item=accessibleClasses().find(x=>x.id===activeClassId);if(!item)return;const w=classWork(item),m=w.members.find(x=>x.id===memberId);if(!m||!confirm(`¿Eliminar definitivamente a ${m.fullName} de esta clase?\n\nSe quitarán también su seguimiento y registros de asistencia dentro de esta clase. Esta acción no se puede deshacer.`))return;w.members=w.members.filter(x=>x.id!==memberId);for(const key of Object.keys(w.tracking))if(key.startsWith(`${memberId}:`))delete w.tracking[key];for(const record of w.attendance){if(record?.entries&&typeof record.entries==='object')delete record.entries[`m:${memberId}`];if(Array.isArray(record?.presentIds))record.presentIds=record.presentIds.filter(id=>String(id)!==String(memberId));}item.updatedAt=isoNow();await persist(true,{action:'Miembro de clase eliminado',entity:item.id,detail:m.fullName});renderClassWorkspace();toast('Conquistador eliminado de la clase.');}
async function resetClassProgress(classId){if(currentUser()?.role!=='admin')return toast('Solo el Administrador puede reiniciar una clase.');const item=(database.classes||[]).find(x=>String(x.id)===String(classId));if(!item)return toast('Clase no encontrada.');const w=classWork(item),members=activeClassMembers(w);if(!confirm(`¿Reiniciar ${item.regular}?\n\nSe pondrán en 0 los avances de todos los Conquistadores, el progreso del plan y la asistencia. Se conservarán los Conquistadores, la planificación, los instructores y los demás datos de la clase.\n\nEsta acción no se puede deshacer.`))return;w.tracking={};for(const plan of Object.values(w.planning||{}))if(plan&&typeof plan==='object')plan.progressed=false;w.attendance=[];item.updatedAt=isoNow();classAttendanceDate=localDate();await persist(true,{action:'Clase reiniciada',entity:item.id,detail:`${members.length} Conquistador(es) · seguimiento, plan y asistencia a 0`});renderClassWorkspace();toast('Clase reiniciada: todos los avances quedaron en 0.');}
async function saveClassAttendanceRecord(id){const item=accessibleClasses().find(x=>x.id===id);if(!item)return;const w=classWork(item),date=classAttendanceDate||localDate();let record=w.attendance.find(r=>r.date===date);if(!record){record={id:newId('att'),date,entries:{},note:''};w.attendance.push(record);}record.note=value('class_att_note').trim();record.entries={};for(const el of document.querySelectorAll('.class-att-status'))record.entries[el.dataset.person]=el.value;delete record.presentIds;await persist(true,{action:'Asistencia registrada',entity:item.id,detail:date});toast('Asistencia guardada.');renderClassWorkspace();}
async function saveClassTrackingCheck(id,memberId,reqId,checked){const item=accessibleClasses().find(x=>x.id===id);if(!item)return;classWork(item).tracking[`${memberId}:${reqId}`]=checked?'completed':'pending';await persist(true,{action:'Avance de Conquistador actualizado',entity:id,detail:`${memberId}:${reqId}:${checked?'completed':'pending'}`});renderClassWorkspace();}

async function saveClubSchedule(file){
  if(!file||currentUser()?.role!=='admin')return;
  const allowed=/^(application\/pdf|image\/)/.test(file.type||'')||/\.(docx?|xlsx?|pptx?)$/i.test(file.name||'');if(!allowed)return toast('Formato no compatible. Usa PDF, imagen, Word, Excel o PowerPoint.');
  const previous=database.settings.clubSchedule;if(previous?.fileId)await fileDelete(previous.fileId);
  const dataUrl=await fileToDataUrl(file),stored={id:newId('club_schedule'),ownerId:currentUser().id,kind:'clubSchedule',name:file.name,type:file.type||'application/octet-stream',size:dataUrlBytes(dataUrl),dataUrl,createdAt:isoNow()};await filePut(stored);
  database.settings.clubSchedule={fileId:stored.id,fileName:file.name,mimeType:stored.type,updatedAt:isoNow(),updatedBy:currentUser().id};await persist(true,{action:'Cronograma del Club actualizado',entity:stored.id,detail:file.name});renderClassWorkspace();toast('Cronograma del Club guardado.');
}
async function openClubSchedule(){
  const meta=database.settings.clubSchedule;if(!meta)return toast('No hay cronograma cargado.');const file=(await allFiles()).find(item=>item.id===meta.fileId);if(!file)return toast('El archivo del cronograma no está disponible en este dispositivo.');
  if((file.type||'').startsWith('image/'))return showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>Cronograma del Club</h3><p>${esc(meta.fileName)}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body preview-image-wrap"><img src="${file.dataUrl}" alt="Cronograma del Club"></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="downloadBlob('${esc(meta.fileName)}',dataUrlToBlob('${file.dataUrl}'))">Guardar copia</button></div></div>`);
  if(file.type==='application/pdf'||/\.pdf$/i.test(file.name||'')){if(materialPreviewUrl)URL.revokeObjectURL(materialPreviewUrl);materialPreviewUrl=URL.createObjectURL(dataUrlToBlob(file.dataUrl));return showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>Cronograma del Club</h3><p>${esc(meta.fileName)}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><iframe class="preview-frame" src="${materialPreviewUrl}#zoom=page-width"></iframe></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="downloadBlob('${esc(meta.fileName)}',dataUrlToBlob('${file.dataUrl}'))">Guardar copia</button></div></div>`);}
  downloadBlob(meta.fileName,dataUrlToBlob(file.dataUrl));toast('El cronograma se guardó para abrirlo con su aplicación correspondiente.');
}
async function removeClubSchedule(){if(currentUser()?.role!=='admin'||!database.settings.clubSchedule)return;const meta=database.settings.clubSchedule;if(!confirm(`¿Retirar el cronograma ${meta.fileName}?`))return;if(meta.fileId)await fileDelete(meta.fileId);delete database.settings.clubSchedule;await persist(true,{action:'Cronograma del Club retirado',entity:'clubSchedule',detail:meta.fileName});renderClassWorkspace();toast('Cronograma retirado.');}
function dataUrlToBlob(dataUrl){const [header,payload]=String(dataUrl||'').split(','),mime=(header.match(/data:([^;]+)/)||[])[1]||'application/octet-stream',binary=atob(payload||''),bytes=Uint8Array.from(binary,ch=>ch.charCodeAt(0));return new Blob([bytes],{type:mime});}


function showAssignedTaskReminder() {
  taskReminderShown=true;
  const user=currentUser();
  if(!user)return;
  const allowed=new Set(accessibleAspirants(user).map(item=>item.id));
  const tasks=(database.tasks||[]).filter(task=>allowed.has(task.assigneeId)&&taskState(task)!=='presented');
  if(!tasks.length)return;
  const overdue=tasks.filter(task=>taskState(task)==='overdue').length;
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>Tareas asignadas</h3><p>Recordatorio al iniciar sesión</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="notice ${overdue?'warn':''}">Tienes <b>${tasks.length}</b> tarea${tasks.length===1?'':'s'} pendiente${tasks.length===1?'':'s'}${overdue?` y <b>${overdue}</b> vencida${overdue===1?'':'s'}`:''}.</div><div class="reading-list" style="margin-top:12px">${tasks.slice(0,6).map(task=>{const req=flatRequirements.find(item=>item.id===task.requirementId);return `<div class="reading-item"><div><h4>${esc(req?.areaTitle||'')} · ${esc(req?.official_text||task.requirementId)}</h4><p>Presentación: ${formatDate(task.dueDate)}</p></div><span class="status ${taskState(task)==='overdue'?'rejected':'assigned'}">${TASK_LABELS[taskState(task)]}</span></div>`;}).join('')}</div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="closeModal();go('tasks')">Ver mis tareas</button></div></div>`);
}

function tutorAspirantMiniCards(tutorId){
  const aspirants=database.users.filter(person=>['aspirant','pastor_aspirant'].includes(person.role)&&sgl545HasTutor(person,tutorId)&&(person.accountState||'active')==='active');
  return `<div class="tutor-aspirant-mini-list">${aspirants.map(aspirant=>`<button class="tutor-aspirant-mini" onclick="activeAspirantId='${aspirant.id}';go('aspirants')"><span class="tutor-mini-photo">${avatarMarkup(aspirant)}</span><span class="tutor-mini-meta"><b>${esc(aspirant.fullName)}</b><small>${progressPercent(aspirant.id)}% Clase Agrupada</small></span><span class="tutor-mini-progress"><i style="width:${progressPercent(aspirant.id)}%"></i></span></button>`).join('')||'<div class="empty compact">Sin Aspirantes asignados.</div>'}</div>`;
}
function renderTutors(){
  const actor=currentUser();
  if(!['admin','regional'].includes(actor?.role))return go('dashboard');
  setTop('Tutores',actor.role==='regional'?'Regional':'Administración');
  const tutors=database.users.filter(item=>item.role==='tutor');
  const canEdit=actor.role==='admin';
  setContent(`<div class="page-head"><div><div class="eyebrow">Revisión de informes</div><h1 class="page-title">Tutores</h1><p class="subtitle">Tutores y Aspirantes asignados, en una vista compacta. Regional puede crear Tutores para Clase Agrupada, Guía Mayor, Guía Mayor Máster y Guía Mayor Máster Avanzado.</p></div><button class="btn primary" onclick="openTutor()">${actor.role==='regional'?'Crear Tutor Regional':'Crear Tutor'}</button></div><div class="tutor-grid-compact">${tutors.map(item=>`<article class="card tutor-card-compact"><div class="tutor-card-head"><span class="avatar tutor-avatar">${avatarMarkup(item)}</span><div><h3>${esc(item.fullName)}</h3><p>${esc(item.email)}</p><span class="chip blue">${database.users.filter(person=>['aspirant','pastor_aspirant'].includes(person.role)&&sgl545HasTutor(person,item.id)).length} Aspirante(s)</span></div>${canEdit?`<button class="btn ghost small" onclick="openUser('${item.id}')">Editar</button>`:'<span class="chip">Solo lectura</span>'}</div>${tutorAspirantMiniCards(item.id)}</article>`).join('')||'<div class="card empty">Todavía no hay Tutores.</div>'}</div>`);
}
function openTutor(id=''){
  openUser(id);
  if(!id){const role=document.getElementById('user_role');if(role){role.value='tutor';toggleUserAssignments();}const title=document.querySelector('.modal-head h3');if(title)title.textContent=currentUser()?.role==='regional'?'Crear Tutor Regional':'Crear Tutor';}
}

function renderUsers() {
  const user = currentUser();
  if (user.role !== 'admin') return go('dashboard');
  setTop('Usuarios y permisos', 'Administración');
  const rows = database.users.map(item => {
    const club = clubFor(item);
    const assigned = ['aspirant','pastor_aspirant'].includes(item.role) ? userById(item.tutorId)?.fullName || '' : '';
    const state=item.accountState||'active';
    return `<tr><td><b>${esc(item.fullName)}</b><br><span class="muted">${esc(item.email || item.username)}</span></td><td><span class="chip gold">${ROLE_LABELS[item.role]||item.role}</span></td><td>${esc(item.region||'—')}</td><td>${esc(club?.name || '—')}</td><td>${esc(assigned || '—')}</td><td><span class="status ${state==='active'?'approved':'rejected'}">${ACCOUNT_LABELS[state]}</span></td><td><div class="actions"><button class="btn ghost small" onclick="openUser('${item.id}')">Editar</button>${item.id!==user.id?`<button class="btn ghost small" onclick="issueTemporaryPassword('${item.id}')">Clave temporal</button>`:''}</div></td></tr>`;
  }).join('');
  setContent(`<div class="page-head"><div><div class="eyebrow">Control exclusivo</div><h1 class="page-title">Cuentas de acceso</h1><p class="subtitle">Solo el Administrador crea Tutores, Aspirantes y Administradores adicionales.</p></div><div class="actions"><button class="btn ghost" onclick="downloadUserTemplate()">Plantilla para Excel</button><label class="btn ghost">Importar CSV<input type="file" accept=".csv,text/csv" hidden onchange="importUsers(this.files[0]);this.value=''"/></label><button class="btn primary" onclick="openUser()">Crear usuario</button></div></div><div class="card table-wrap"><table class="table"><thead><tr><th>Persona</th><th>Rol</th><th>Región</th><th>Club</th><th>Tutor</th><th>Estado</th><th></th></tr></thead><tbody>${rows || '<tr><td colspan="6" class="empty">No hay usuarios.</td></tr>'}</tbody></table></div><div class="notice" style="margin-top:14px">Primero se inactiva o deshabilita una cuenta sin borrar sus datos. El Administrador puede reactivarla o archivarla después.</div>`);
}

function openUser(id='') {
  const editor = id ? userById(id) : null;
  const actor=currentUser();
  if(actor?.role==='tutor'&&(id||editor))return toast('El Tutor puede crear Aspirantes; la edición de cuentas corresponde al Administrador.');
  if(actor?.role==='regional'&&id)return toast('Regional puede crear Tutores, pero no editar cuentas existentes.');
  if(!['admin','tutor','regional','lider_musica'].includes(actor?.role))return;
  const tutorCreating=actor.role==='tutor';
  const regionalCreatingTutor=actor.role==='regional'&&!id;
  const tutors = database.users.filter(item => item.role === 'tutor' && item.active !== false);
  const clubOptions = `<option value="">Sin Club</option>${database.clubs.filter(item => item.active !== false).map(club => `<option value="${club.id}" ${editor?.clubId === club.id ? 'selected' : ''}>${esc(club.name)}</option>`).join('')}`;
  const roles=tutorCreating?[['aspirant','Aspirante']]:(regionalCreatingTutor?[['tutor','Tutor']]:Object.entries(ROLE_LABELS));
  const tutorOptions=tutorCreating?`<option value="${actor.id}" selected>${esc(actor.fullName)}</option>`:`<option value="">Sin asignar</option>${tutors.map(item => `<option value="${item.id}" ${editor?.tutorId === item.id ? 'selected' : ''}>${esc(item.fullName)}</option>`).join('')}`;
  const photoEditor=actor.role==='admin'&&editor?`<div class="field-group full"><span>Foto de perfil</span><div class="user-row"><div class="avatar">${avatarMarkup(editor)}</div><div class="actions"><button type="button" class="btn ghost small" onclick="openUserPhotoStudio('${editor.id}')">Editar foto en estudio</button><label class="btn ghost small">Reemplazar rápido<input id="user_photo" type="file" hidden accept="image/*"></label></div></div></div>`:actor.role==='admin'?'<label class="field-group full">Foto de perfil opcional<input id="user_photo" type="file" class="field" accept="image/*"></label>':'';
  showModal(`<div class="modal user-editor"><div class="modal-head"><div><h3>${editor ? 'Editar usuario' : (regionalCreatingTutor?'Crear Tutor Regional':'Crear usuario')}</h3><p>${regionalCreatingTutor?'Regional crea Tutores para Clase Agrupada, Guía Mayor, Guía Mayor Máster y Guía Mayor Máster Avanzado.':(tutorCreating?'El nuevo Aspirante quedará asignado automáticamente a tu tutoría.':'El Administrador define el rol, datos institucionales y asignaciones.')}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveUser(event,'${id}')"><div class="modal-body"><div class="field-row"><label class="field-group full">Nombres y Apellidos Completos<input id="user_name" class="field" required autocomplete="off" list="user_name_history" value="${esc(editor?.fullName || '')}"></label><datalist id="user_name_history">${nameAutocompleteOptions()}</datalist><div class="field-group full"><button type="button" class="btn link small" onclick="clearNameAutocompleteHistory()">Limpiar nombres guardados</button></div><label class="field-group">Correo<input id="user_email" class="field" type="email" required ${editor&&['aspirant','pastor_aspirant'].includes(editor.role)?'readonly':''} value="${esc(editor?.email || '')}"></label><label class="field-group">Usuario<input id="user_username" class="field" required value="${esc(editor?.username || '')}"></label><label class="field-group">Fecha de nacimiento<input id="user_birth" class="field" type="date" value="${esc(editor?.birthDate||'')}"></label><label class="field-group">Teléfono celular<input id="user_phone" class="field" type="tel" autocomplete="tel" value="${esc(editor?.phone||'')}"></label><label class="field-group">Rol<select id="user_role" class="field" onchange="toggleUserAssignments()">${roles.map(([role,label]) => `<option value="${role}" ${(editor?.role||'aspirant') === role ? 'selected' : ''}>${label}</option>`).join('')}</select></label><label class="field-group">Estado<select id="user_state" class="field">${Object.entries(ACCOUNT_LABELS).map(([state,label])=>`<option value="${state}" ${(editor?.accountState||'active')===state?'selected':''}>${label}</option>`).join('')}</select></label><label class="field-group">Club<select id="user_club" class="field">${clubOptions}</select></label><label id="user_assignments" class="field-group">Tutor asignado<select id="user_tutor" class="field">${tutorOptions}</select></label><label class="field-group">Iglesia a la que asiste<input id="user_church" class="field" value="${esc(editor?.church||'')}"></label><label class="field-group">Cargo en el Club<input id="user_club_position" class="field" value="${esc(editor?.clubPosition||'')}"></label><label class="field-group">Cargo en la Iglesia<input id="user_church_position" class="field" value="${esc(editor?.churchPosition||'')}"></label><label class="field-group">Distrito Misionero<input id="user_district" class="field" value="${esc(editor?.missionaryDistrict||'')}"></label><label class="field-group">Región<input id="user_region" class="field" placeholder="1, 2, 3…" value="${esc(editor?.region||'')}"></label><label class="field-group">Misión<input id="user_mission" class="field" value="${esc(editor?.mission||'')}"></label><label class="field-group">Unión<input id="user_union" class="field" value="${esc(editor?.union||'')}"></label>${photoEditor}<label class="field-group full">${editor ? 'Nueva contraseña (dejar vacía para conservar)' : 'Contraseña temporal'}<input id="user_password" type="password" minlength="8" class="field" ${editor ? '' : 'required'}></label></div><div id="user_error" class="error"></div></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar usuario</button></div></form></div>`);
  toggleUserAssignments();
}

function toggleUserAssignments() {
  const role = value('user_role');
  document.getElementById('user_assignments')?.classList.toggle('hidden', !['aspirant','pastor_aspirant'].includes(role));
}

async function saveUser(event, id='') {
  event.preventDefault();
  const actor=currentUser(),tutorCreating=actor?.role==='tutor'&&!id,regionalCreatingTutor=actor?.role==='regional'&&!id;
  if (actor?.role !== 'admin'&&!tutorCreating&&!regionalCreatingTutor) return;
  const email = normalize(value('user_email'));
  const username = normalize(value('user_username'));
  const duplicate = database.users.find(item => item.id !== id && (normalize(item.email) === email || normalize(item.username) === username));
  if (duplicate) return setText('user_error', 'El correo o usuario ya pertenece a otra cuenta.');
  const password = value('user_password');
  const item = id ? userById(id) : {id:newId('usr'), createdAt:isoNow(), photo:'', drive:null};
  item.theme ||= 'classic_blue';
  item.fullName = value('user_name').trim();
  item.email = id&&['aspirant','pastor_aspirant'].includes(item.role)?item.email:email;
  item.username = username;
  item.birthDate=value('user_birth');item.phone=value('user_phone').trim();
  item.role = regionalCreatingTutor?'tutor':(tutorCreating?'aspirant':value('user_role'));
  if(regionalCreatingTutor){item.regionalCreatedBy=actor.id;item.regionalProgramIds=[...REGIONAL_PROGRAM_IDS];}
  item.clubId = value('user_club') || null;
  item.accountState=value('user_state')||'active'; item.active=item.accountState==='active';
  item.tutorId = ['aspirant','pastor_aspirant'].includes(item.role) ? (tutorCreating?actor.id:value('user_tutor') || null) : null;
  item.reportFont ||= database.settings.reportFont||'Arial';
  item.church=value('user_church').trim(); item.clubPosition=value('user_club_position').trim(); item.churchPosition=value('user_church_position').trim();
  item.missionaryDistrict=value('user_district').trim(); item.region=value('user_region').trim(); item.mission=value('user_mission').trim(); item.union=value('user_union').trim();
  const photo=document.getElementById('user_photo')?.files?.[0];if(photo)item.photo=await makeSquareProfilePhoto(photo);
  if (password) item.password = await passwordRecord(password);
  if (!id) database.users.push(item);
  if (['aspirant','pastor_aspirant'].includes(item.role)) ensureProgress(item.id);
  rememberAutocompleteName(item.fullName);
  await persist(true, {action:id ? 'Usuario actualizado' : 'Usuario creado', entity:item.id, detail:`${item.fullName} · ${ROLE_LABELS[item.role]}`});
  closeModal();
  if(['admin','regional'].includes(actor.role))currentPage==='tutors'?renderTutors():renderUsers();else renderAspirants();
  toast('Usuario guardado.');
}

function downloadUserTemplate() {
  downloadText('Plantilla_Usuarios_MinisterioMusica.csv', '\ufeffnombre,correo,usuario,fecha_nacimiento,telefono,rol,club,tutor_correo,estado,iglesia,cargo_club,cargo_iglesia,distrito_misionero,region,mision,union,contrasena_temporal\nNombre del Aspirante,aspirante@ejemplo.com,aspirante,2000-01-31,+59170000000,aspirant,Nombre del Club,tutor@ejemplo.com,active,Iglesia,Cargo,Cargo,Distrito,1,Misión,Unión,Cambiar123', 'text/csv;charset=utf-8');
}

async function issueTemporaryPassword(id) { if(currentUser()?.role!=='admin')return; const user=userById(id); if(!user)return; const temporary=`CA-${Math.random().toString(36).slice(2,7)}-${Math.random().toString(36).slice(2,7)}!`; user.password=await passwordRecord(temporary); user.mustChangePassword=true; await persist(true,{action:'Contraseña temporal emitida',entity:user.id,detail:user.fullName}); prompt(`Contraseña temporal para ${user.fullName}. Cópiala y entrégala de forma segura:`,temporary); }

async function importUsers(file) {
  if (!file || currentUser()?.role !== 'admin') return;
  const rows = parseCsv(await file.text());
  if (rows.length < 2) return toast('El archivo CSV no contiene filas.');
  const headers = rows[0].map(normalize);
  let created = 0, skipped = 0;
  for (const cells of rows.slice(1)) {
    const record = Object.fromEntries(headers.map((header,index) => [header, String(cells[index] || '').trim()]));
    if (!record.nombre || !record.correo) { skipped += 1; continue; }
    const roleAliases = {administrador:'admin',admin:'admin',tutor:'tutor',lider:'aspirant','líder':'aspirant','lider aspirante':'aspirant','líder aspirante':'aspirant',aspirante:'aspirant',leader:'aspirant',aspirant:'aspirant'};
    const role = roleAliases[normalize(record.rol)];
    if (!role || database.users.some(item => normalize(item.email) === normalize(record.correo) || normalize(item.username) === normalize(record.usuario))) { skipped += 1; continue; }
    const club = database.clubs.find(item => normalize(item.name) === normalize(record.club));
    const tutor = database.users.find(item => item.role === 'tutor' && normalize(item.email) === normalize(record.tutor_correo));
    const password = record.contrasena_temporal?.length >= 8 ? record.contrasena_temporal : `Clase${Math.random().toString(36).slice(2,8)}!`;
    const accountState=['inactive','archived'].includes(normalize(record.estado))?normalize(record.estado):'active';
    const item = {id:newId('usr'), fullName:record.nombre, email:normalize(record.correo), username:normalize(record.usuario || record.correo.split('@')[0]), birthDate:record.fecha_nacimiento||'',phone:record.telefono||'',role, clubId:club?.id || null, tutorId:['aspirant','pastor_aspirant'].includes(role) ? tutor?.id || null : null, accountState, active:accountState==='active', church:record.iglesia||'',clubPosition:record.cargo_club||'',churchPosition:record.cargo_iglesia||'',missionaryDistrict:record.distrito_misionero||'',region:record.region||'',mission:record.mision||'',union:record.union||'',reportFont:'Arial',theme:'classic_blue',createdAt:isoNow(), photo:'', password:await passwordRecord(password)};
    database.users.push(item);
    if (['aspirant','pastor_aspirant'].includes(role)) ensureProgress(item.id);
    created += 1;
  }
  await persist(true, {action:'Usuarios importados', entity:'users', detail:`${created} creados; ${skipped} omitidos`});
  renderUsers();
  toast(`${created} usuarios creados; ${skipped} filas omitidas.`);
}

function parseCsv(text) {
  const output=[]; let row=[], cell='', quoted=false;
  for (let i=0;i<text.length;i+=1) { const char=text[i], next=text[i+1]; if (char==='"' && quoted && next==='"') { cell+='"'; i+=1; } else if (char==='"') quoted=!quoted; else if (char===',' && !quoted) { row.push(cell); cell=''; } else if ((char==='\n' || char==='\r') && !quoted) { if (char==='\r' && next==='\n') i+=1; row.push(cell); if (row.some(value => value.trim())) output.push(row); row=[]; cell=''; } else cell+=char; }
  row.push(cell); if (row.some(value => value.trim())) output.push(row); return output;
}

function renderClubs() {
  setTop('Clubes', 'Administración');
  const cards = database.clubs.map(club => `<div class="card club-card"><img class="club-logo" src="${club.logo || appLogo()}"><div><h3>${esc(club.name)}</h3><p class="subtitle">${esc([club.church,club.missionaryDistrict,club.region].filter(Boolean).join(' · ') || 'Ubicación no indicada')}</p><span class="chip">${database.users.filter(item => item.clubId === club.id).length} miembros</span>${club.mapsUrl?`<a class="btn link small" target="_blank" rel="noopener" href="${esc(club.mapsUrl)}">Abrir ubicación en Maps</a>`:''}</div><div class="actions"><button class="btn ghost small" onclick="openClub('${club.id}')">Editar</button></div></div>`).join('');
  setContent(`<div class="page-head"><div><div class="eyebrow">Identidad opcional</div><h1 class="page-title">Clubes</h1><p class="subtitle">El triángulo de Conquistadores siempre permanece; el nombre y logo del Club son opcionales.</p></div><button class="btn primary" onclick="openClub()">Crear Club</button></div><div class="grid3">${cards || '<div class="card empty">Todavía no hay Clubes.</div>'}</div>`);
}

function openClub(id='') {
  const club = database.clubs.find(item => item.id === id);
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>${club ? 'Editar Club' : 'Crear Club'}</h3><p>Estos datos se muestran a sus miembros y en los informes.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveClub(event,'${id}')"><div class="modal-body"><label class="field-group">Nombre del Club<input id="club_name" class="field" required value="${esc(club?.name || '')}"></label><label class="field-group">Iglesia<input id="club_church" class="field" value="${esc(club?.church || '')}"></label><label class="field-group">Distrito Misionero<input id="club_district" class="field" value="${esc(club?.missionaryDistrict || '')}"></label><label class="field-group">Región<select id="club_region" class="field"><option value="">Seleccionar región</option>${[...new Set([...Array.from({length:12},(_,i)=>String(i+1)),'MBOS',club?.region||''])].filter(Boolean).map(r=>`<option value="${esc(r)}" ${String(club?.region||'')===r?'selected':''}>${esc(r)}</option>`).join('')}</select></label><label class="field-group">Misión<input id="club_mission" class="field" value="${esc(club?.mission || '')}"></label><label class="field-group">Unión<input id="club_union" class="field" value="${esc(club?.union || '')}"></label><label class="field-group">Ubicación de Google Maps<input id="club_maps" type="url" class="field" placeholder="https://maps.google.com/…" value="${esc(club?.mapsUrl || '')}"></label><div class="user-row sgl562-club-logo"><img id="club_logo_preview" class="club-logo" src="${esc(club?.logo||appLogo())}" alt="Logo del club"><button type="button" class="btn ghost" onclick="openClubLogoStudio()">Editar logo en estudio</button></div><label class="field-group">Logo opcional<input id="club_logo" class="field" type="file" accept="image/*" onchange="delete this.sglLogoDraft"></label><div class="notice">Toda imagen se autoajusta. Si no carga logo, el informe mostrará el nombre del Club junto al nombre completo.</div></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar Club</button></div></form></div>`);
}

async function saveClub(event,id='') {
  event.preventDefault();
  const club = database.clubs.find(item => item.id === id) || {id:newId('club'), createdAt:isoNow(), active:true, logo:''};
  club.name=value('club_name').trim(); club.church=value('club_church').trim(); club.missionaryDistrict=value('club_district').trim(); club.region=value('club_region').trim(); club.mission=value('club_mission').trim(); club.union=value('club_union').trim(); club.mapsUrl=value('club_maps').trim();
  const input=document.getElementById('club_logo'),file=input.files[0];if(Object.hasOwn(input,'sglLogoDraft'))club.logo=input.sglLogoDraft;else if(file)club.logo=await imageToTransparentDataUrl(file,1200);
  if (!id) database.clubs.push(club);
  await persist(true,{action:id?'Club actualizado':'Club creado',entity:club.id,detail:club.name}); closeModal(); renderClubs(); toast('Club guardado.');
}

function renderAspirants() {
  const user=currentUser(), people=accessibleAspirants(user);
  setTop(user.role==='admin'?'Aspirantes':'Mis Aspirantes',ROLE_LABELS[user.rawRole624||user.role]);
  const cards=people.map(person=>`<div class="card"><div class="card-head"><div class="user-row"><div class="avatar">${avatarMarkup(person)}</div><div><h3 style="margin:0">${esc(person.fullName)}</h3><span class="muted" style="font-size:11px">${esc(clubFor(person)?.name || 'Sin Club')}</span></div></div><span class="status approved">${progressPercent(person.id)}%</span></div><div class="profile-facts">${person.birthDate?`<span>🎂 ${String(person.birthDate).slice(5).split('-').reverse().join('/')}</span>`:''}${person.phone?`<span>📱 ${esc(person.phone)}</span>`:''}</div><div class="bar"><i style="width:${progressPercent(person.id)}%"></i></div><p class="muted" style="font-size:11px">${countStatus(person.id,'approved')} revisados · ${countStatus(person.id,'in_review')} en revisión</p><div class="actions"><button class="btn dark block" onclick="openAspirant('${person.id}')">Abrir expediente</button>${user.role==='admin'?`<button class="btn ghost block" onclick="openUser('${person.id}')">Editar perfil</button>`:''}</div></div>`).join('');
  setContent(`<div class="page-head"><div><div class="eyebrow">Seguimiento autorizado</div><h1 class="page-title">${people.length} Aspirante${people.length===1?'':'s'}</h1><p class="subtitle">Cada Tutor ve únicamente las personas que tiene asignadas.</p></div><div class="actions">${['admin','tutor','lider_musica'].includes(user.role)?'<button class="btn primary" onclick="openUser()">Crear Aspirante</button>':''}<button class="btn ghost" onclick="exportAspirantsPdf()">Exportar PDF</button></div></div><div class="grid3">${cards || '<div class="card empty">No hay Aspirantes asignados.</div>'}</div>`);
}

function exportAspirantsCsv() { const people=accessibleAspirants(),header='nombre,correo,usuario,club,tutor,revisados,total,porcentaje\n',rows=people.map(item=>[item.fullName,item.email,item.username,clubFor(item)?.name||'',userById(item.tutorId)?.fullName||'',countStatus(item.id,'approved'),flatRequirements.length,progressPercent(item.id)].map(csvCell).join(',')).join('\n'); downloadText(`Aspirantes_${localDate()}.csv`,`\ufeff${header}${rows}`,'text/csv;charset=utf-8'); }
async function exportAspirantsPdf(){const people=accessibleAspirants(),blob=makeAspirantsListPdf(people),result=await saveBlobWithDialog(`Aspirantes_${localDate()}.pdf`,blob,'Documento PDF|*.pdf');if(result?.ok)toast(`Listado guardado en ${result.path}.`);}
function makeAspirantsListPdf(people){const pages=[];let page,y;const addPage=()=>{page={w:612,h:792,commands:[],images:[]};pages.push(page);pdfText(page,'SISTEMA DE GESTIÓN DE MÚSICA',50,48,16,'AB');pdfText(page,'LISTADO DE ASPIRANTES',50,70,11,'AB');pdfText(page,`Fecha: ${formatDate(localDate())}`,430,70,9,'AR');page.commands.push('0.714 0.251 0.251 RG 50 700 m 562 700 l 2 w S 0 0 0 RG');y=108;drawHead();};const drawHead=()=>{page.commands.push(`0.96 0.91 0.72 rg 50 ${(page.h-y-20).toFixed(2)} 512 20 re f 0 0 0 rg`);[['Aspirante',54],['Iglesia',238],['Tutor',352],['Avance',510]].forEach(([text,x])=>pdfText(page,text,x,y+4,9,'AB'));y+=24;};addPage();for(const person of people){if(y+34>page.h-55){addPage();}const fill=people.indexOf(person)%2?'0.98 0.98 0.98':'1 1 1';page.commands.push(`${fill} rg 50 ${(page.h-y-30).toFixed(2)} 512 30 re f 0 0 0 rg`);pdfText(page,short(person.fullName,31),54,y+5,9,'AR');pdfText(page,short(clubFor(person)?.name||'Sin Club',18),238,y+5,9,'AR');pdfText(page,short(userById(person.tutorId)?.fullName||'Sin Tutor',22),352,y+5,9,'AR');pdfText(page,`${countStatus(person.id,'approved')}/${flatRequirements.length} · ${progressPercent(person.id)}%`,510,y+5,9,'AB');page.commands.push(`0.88 0.88 0.88 RG 50 ${(page.h-y-30).toFixed(2)} m 562 ${(page.h-y-30).toFixed(2)} l .5 w S 0 0 0 RG`);y+=30;}return buildVectorPdf(pages);}
function csvCell(valueText) { const textValue=String(valueText??''); return /[",\n]/.test(textValue)?`"${textValue.replace(/"/g,'""')}"`:textValue; }

function openAspirant(id) { activeAspirantId=id; go('requirements'); }

function taskState(task) { if(task.presentedAt)return 'presented'; return task.dueDate&&task.dueDate<localDate()?'overdue':'pending'; }
function taskRequirementLabel(req){return{area:req?.areaTitle||'Área no indicada',id:req?.id||'—',short:compactRequirementText(req||{})};}
function renderTasks() {
  const user=currentUser(), canAssign=['admin','tutor','lider_musica'].includes(user.role), people=accessibleAspirants(user);
  const allowedIds=new Set(['aspirant','pastor_aspirant'].includes(user.role)?[user.id]:people.map(item=>item.id));
  const tasks=(database.tasks||[]).filter(task=>allowedIds.has(task.assigneeId)).sort((a,b)=>String(a.dueDate).localeCompare(String(b.dueDate)));
  setTop(['aspirant','pastor_aspirant'].includes(user.role)?'Mis tareas':'Asignación de tareas',ROLE_LABELS[user.rawRole624||user.role]);
  setContent(`<div class="page-head"><div><div class="eyebrow">Fechas de presentación</div><h1 class="page-title">${tasks.length} tarea${tasks.length===1?'':'s'}</h1><p class="subtitle">Estados automáticos: Por presentar, Presentado o Vencido.</p></div>${canAssign?'<button class="btn primary" onclick="openTask()">Asignar requisito</button>':''}</div><div class="card table-wrap"><table class="table"><thead><tr><th>Aspirante</th><th>Requisito</th><th>Presentación</th><th>Estado</th><th>Indicaciones</th><th></th></tr></thead><tbody>${tasks.map(task=>{const state=taskState(task),req=flatRequirements.find(item=>item.id===task.requirementId);return `<tr><td>${esc(userById(task.assigneeId)?.fullName||'—')}</td><td>${(()=>{const label=taskRequirementLabel(req);return `<div class="task-requirement"><b>${esc(req?.areaTitle||label.area)}</b><span>${esc(label.id)} · ${esc(label.short)}</span></div>`;})()}</td><td>${formatDate(task.dueDate)}</td><td><span class="status ${state==='presented'?'approved':state==='overdue'?'rejected':'assigned'}">${TASK_LABELS[state]}</span></td><td>${esc(short(task.instructions||'',70))}</td><td>${canAssign?`<button class="btn ghost small" onclick="openTask('${task.id}')">Editar</button>`:`<button class="btn ghost small" onclick="openRequirement('${task.requirementId}')">Abrir</button>`}</td></tr>`;}).join('')||'<tr><td colspan="6" class="empty">No hay tareas asignadas.</td></tr>'}</tbody></table></div>`);
}

function openTask(id='') {
  if(!['admin','tutor','lider_musica'].includes(currentUser()?.role))return; const task=(database.tasks||[]).find(item=>item.id===id),people=accessibleAspirants();
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>${task?'Editar tarea':'Asignar requisito'}</h3><p>Define Aspirante y fecha de presentación.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveTask(event,'${id}')"><div class="modal-body"><label class="field-group">Aspirante<select id="task_assignee" class="field" required>${people.map(person=>`<option value="${person.id}" ${task?.assigneeId===person.id?'selected':''}>${esc(person.fullName)}</option>`).join('')}</select></label><label class="field-group">Requisito<select id="task_requirement" class="field" required>${flatRequirements.map(req=>{const label=taskRequirementLabel(req);return `<option value="${req.id}" ${task?.requirementId===req.id?'selected':''}>${esc(req.areaTitle||label.area)} — ${esc(label.id)} · ${esc(label.short)}</option>`;}).join('')}</select></label><label class="field-group">Fecha de presentación<input id="task_due" type="date" class="field" required value="${esc(task?.dueDate||localDate())}"></label><label class="field-group">Indicaciones<textarea id="task_instructions" class="field">${esc(task?.instructions||'')}</textarea></label></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar tarea</button></div></form></div>`);
}

async function saveTask(event,id='') {
  event.preventDefault(); if(!['admin','tutor','lider_musica'].includes(currentUser()?.role))return; const assigneeId=value('task_assignee');
  if(!accessibleAspirants().some(item=>item.id===assigneeId))return toast('No tienes permiso para asignar a esa persona.');
  const task=(database.tasks||[]).find(item=>item.id===id)||{id:newId('task'),createdAt:isoNow(),createdBy:currentUser().id,presentedAt:null};
  task.assigneeId=assigneeId; task.requirementId=value('task_requirement'); task.dueDate=value('task_due'); task.instructions=value('task_instructions').trim(); task.updatedAt=isoNow();
  if(!id)database.tasks.push(task); const progress=ensureProgress(assigneeId); if(progress.status[task.requirementId]==='pending')progress.status[task.requirementId]='assigned';
  await persist(true,{action:id?'Tarea actualizada':'Tarea asignada',entity:task.id,detail:`${task.requirementId} · ${userById(assigneeId)?.fullName}`}); closeModal(); renderTasks(); toast('Tarea guardada.');
}

function accessibleMaterials(user=currentUser()){
  if(!user)return[];
  const materials=database.materials||[];
  if(user.role==='admin')return materials;
  if(user.role==='tutor')return materials.filter(item=>(item.active!==false&&userById(item.createdBy)?.role==='admin')||item.createdBy===user.id);
  return materials.filter(item=>item.active!==false&&(userById(item.createdBy)?.role==='admin'||item.createdBy===user.tutorId));
}
function requirementMaterials(requirementId,user=currentUser()){return accessibleMaterials(user).filter(item=>item.active!==false&&item.requirementId===requirementId);}
function renderMaterials(){const user=currentUser(),items=accessibleMaterials(user),canManage=['admin','tutor','lider_musica'].includes(user.role),category=value('material_filter');setTop('Materiales de apoyo',ROLE_LABELS[user.rawRole624||user.role]);const filtered=items.filter(item=>!category||item.category===category);setContent(`<div class="page-head"><div><div class="eyebrow">Biblioteca compartida</div><h1 class="page-title">Materiales</h1><p class="subtitle">PDF, Word, Excel y videos, generales o asociados a un requisito.</p></div>${canManage?'<button class="btn primary" onclick="openMaterial()">Agregar material</button>':''}</div><div class="material-folders"><button class="material-folder ${!category?'active':''}" onclick="document.getElementById('material_filter').value='';renderMaterials()"><b>▦</b><span>Todos</span><small>${items.length}</small></button>${Object.entries(MATERIAL_LABELS).map(([key,label])=>`<button class="material-folder ${category===key?'active':''}" onclick="document.getElementById('material_filter').value='${key}';renderMaterials()"><b>${key==='videos'?'▶':'▤'}</b><span>${label}</span><small>${items.filter(item=>item.category===key).length}</small></button>`).join('')}</div><input id="material_filter" type="hidden" value="${esc(category)}"><div class="list" style="margin-top:15px">${filtered.map(item=>{const req=flatRequirements.find(req=>req.id===item.requirementId),owner=userById(item.createdBy),editable=canManage&&(user.role==='admin'||item.createdBy===user.id);return `<div class="list-item"><div class="req-code">${materialIcon(item)}</div><div><h4>${esc(item.title)}</h4><div class="chips"><span class="chip blue">${esc(MATERIAL_LABELS[item.category]||'Material')}</span>${req?`<span class="chip blue">${esc(req.areaTitle)} · ${esc(req.id)}</span>`:'<span class="chip">Material general</span>'}<span class="chip">${esc(owner?.fullName||'Administrador')}</span>${item.active===false?'<span class="chip red">No disponible</span>':''}</div>${item.description?`<p class="muted" style="font-size:11px;margin:7px 0 0">${esc(item.description)}</p>`:''}</div><div class="actions">${item.active!==false?`<button class="btn primary small" onclick="previewMaterial('${item.id}')">Visualizar</button>${item.url?'':`<button class="btn ghost small" onclick="downloadMaterial('${item.id}')">Guardar</button>`}`:''}${editable?`<button class="btn ghost small" onclick="openMaterial('${item.id}')">Editar</button>`:''}</div></div>`;}).join('')||'<div class="card empty">No hay materiales en esta carpeta.</div>'}</div>`);}
function openMaterial(id=''){const user=currentUser();if(!['admin','tutor','lider_musica'].includes(user?.role))return;const item=(database.materials||[]).find(material=>material.id===id);if(item&&user.role!=='admin'&&item.createdBy!==user.id)return;showModal(`<div class="modal small"><div class="modal-head"><div><h3>${item?'Editar material':'Agregar material'}</h3><p>Puede ser general o estar asociado a un requisito.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveMaterial(event,'${id}')"><div class="modal-body"><label class="field-group">Título<input id="material_title" class="field" required value="${esc(item?.title||'')}"></label><label class="field-group">Carpeta<select id="material_category" class="field">${Object.entries(MATERIAL_LABELS).map(([key,label])=>`<option value="${key}" ${item?.category===key?'selected':''}>${label}</option>`).join('')}</select></label><label class="field-group">Asociar a requisito (opcional)<select id="material_requirement" class="field"><option value="">Sin requisito · material general</option>${flatRequirements.map(req=>`<option value="${req.id}" ${item?.requirementId===req.id?'selected':''}>${esc(req.areaTitle)} · ${esc(short(req.official_text,72))}</option>`).join('')}</select></label><label class="field-group">Descripción<textarea id="material_description" class="field" spellcheck="true">${esc(item?.description||'')}</textarea></label><label class="field-group">Archivo PDF, Word, Excel o video<input id="material_file" type="file" class="field" accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.mp4,.webm,.mov,.avi,image/*,video/*,application/pdf"></label><div class="divider">o enlace</div><label class="field-group">Enlace de video o recurso<input id="material_url" type="url" class="field" placeholder="https://…" value="${esc(item?.url||'')}"></label>${item?'<label class="choice"><input id="material_active" type="checkbox" '+(item.active!==false?'checked':'')+'> Material disponible</label>':''}<div id="material_error" class="error"></div></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar material</button></div></form></div>`);}
async function saveMaterial(event,id=''){event.preventDefault();const user=currentUser();if(!['admin','tutor','lider_musica'].includes(user?.role))return;const item=(database.materials||[]).find(material=>material.id===id)||{id:newId('mat'),createdAt:isoNow(),createdBy:user.id,active:true,fileId:null};if(id&&user.role!=='admin'&&item.createdBy!==user.id)return;const file=document.getElementById('material_file')?.files?.[0],url=value('material_url').trim();if(url&&!/^https?:\/\//i.test(url))return setText('material_error','El enlace debe comenzar con http:// o https://.');if(!file&&!url&&!item.fileId&&!item.dataUrl)return setText('material_error','Carga un archivo o escribe un enlace.');if(file){if(file.size>30*1024*1024)return setText('material_error','El archivo supera 30 MB. Usa un enlace para materiales más grandes.');const fileId=item.fileId||newId('file'),dataUrl=await fileToDataUrl(file);await filePut({id:fileId,ownerId:user.id,kind:'support_material',name:file.name,type:file.type||'application/octet-stream',size:file.size,dataUrl,createdAt:isoNow()});item.fileId=fileId;item.fileName=file.name;item.fileType=file.type||'application/octet-stream';delete item.dataUrl;item.url='';}else if(url){item.url=url;delete item.dataUrl;}item.title=value('material_title').trim();item.category=value('material_category');item.requirementId=value('material_requirement')||null;item.description=value('material_description').trim();item.updatedAt=isoNow();if(id)item.active=document.getElementById('material_active')?.checked!==false;else database.materials.push(item);await persist(true,{action:id?'Material actualizado':'Material publicado',entity:item.id,detail:item.title});closeModal();renderMaterials();toast('Material guardado.');}
async function downloadMaterial(id){const item=accessibleMaterials().find(material=>material.id===id);if(!item)return;if(item.url){window.open(item.url,'_blank','noopener');return;}const files=await allFiles();const file=files.find(entry=>entry.id===item.fileId);const source=file?.dataUrl||item.dataUrl;if(!source){return toast('El material existe, pero el archivo PDF no llegó al dispositivo del líder. Sincroniza los archivos adjuntos y vuelve a intentar.');}const blob=await (await fetch(source)).blob(),result=await saveBlobWithDialog(file?.name||item.fileName||'Material',blob,'Todos los archivos|*.*');if(result?.ok)toast(`Material guardado en ${result.path}.`);}
function materialIcon(item){const type=String(item.fileType||'').toLowerCase(),name=String(item.fileName||'').toLowerCase();if(item.category==='videos'||type.startsWith('video/'))return '▶';if(type.includes('pdf')||name.endsWith('.pdf'))return 'PDF';if(/\.docx?$/.test(name))return 'W';if(/\.xlsx?$/.test(name))return 'X';return 'DOC';}
function videoEmbedUrl(url){try{const parsed=new URL(url);if(/(^|\.)youtube\.com$/i.test(parsed.hostname)){const id=parsed.searchParams.get('v');return id?`https://www.youtube.com/embed/${encodeURIComponent(id)}`:'';}if(/(^|\.)youtu\.be$/i.test(parsed.hostname))return `https://www.youtube.com/embed/${encodeURIComponent(parsed.pathname.slice(1))}`;if(/(^|\.)vimeo\.com$/i.test(parsed.hostname))return `https://player.vimeo.com/video/${encodeURIComponent(parsed.pathname.split('/').filter(Boolean).at(-1)||'')}`;}catch{}return '';}
async function previewMaterial(id){
  const item=accessibleMaterials().find(material=>material.id===id);if(!item)return;
  if(item.url){const embed=videoEmbedUrl(item.url);if(embed){showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>${esc(item.title)}</h3><p>Vista previa de video</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><iframe class="preview-frame" src="${esc(embed)}" allow="accelerometer; autoplay; encrypted-media; picture-in-picture" allowfullscreen></iframe></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="window.open('${esc(item.url)}','_blank','noopener')">Abrir enlace</button></div></div>`);return;}window.open(item.url,'_blank','noopener');return;}
  const file=(await allFiles()).find(entry=>entry.id===item.fileId),source=file?.dataUrl||item.dataUrl;if(!source)return toast('El material está registrado, pero falta el archivo adjunto en este dispositivo. Sincroniza archivos y vuelve a intentar.');
  const name=file?.name||item.fileName||'Material',type=String(file?.type||item.fileType||'').toLowerCase();
  if(type==='application/pdf'||name.toLowerCase().endsWith('.pdf')){const blob=await (await fetch(source)).blob();materialPreviewUrl=URL.createObjectURL(blob);showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>${esc(item.title)}</h3><p>Documento PDF</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><iframe class="preview-frame" src="${materialPreviewUrl}#zoom=page-width"></iframe></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="downloadMaterial('${item.id}')">Guardar una copia</button></div></div>`);return;}
  if(type.startsWith('video/')){showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>${esc(item.title)}</h3><p>Reproductor de video</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><video class="preview-video" src="${source}" controls></video></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="downloadMaterial('${item.id}')">Guardar una copia</button></div></div>`);return;}
  if(type.startsWith('image/')){showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>${esc(item.title)}</h3><p>Vista previa de imagen</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body preview-image-wrap"><img src="${source}" alt="${esc(item.title)}"></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="downloadMaterial('${item.id}')">Guardar una copia</button></div></div>`);return;}
  try{const blob=await (await fetch(source)).blob(),result=await nativeRequest('/file/open',{filename:name,dataBase64:await blobToBase64(blob)});if(result?.ok)toast('Material abierto con el programa instalado en Windows.');else toast(result?.error||'No se pudo abrir el material.');}catch{toast('No se encontró un visor local. Puedes guardar el archivo y abrirlo con Word o Excel.');}
}
function openRequirementMaterials(requirementId){const items=requirementMaterials(requirementId);showModal(`<div class="modal small"><div class="modal-head"><div><h3>Material de apoyo</h3><p>${esc(flatRequirements.find(item=>item.id===requirementId)?.official_text||requirementId)}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="reading-list">${items.map(item=>`<div class="reading-item"><div><h4>${esc(item.title)}</h4><p>${esc(MATERIAL_LABELS[item.category]||'Material')} · ${esc(item.description||'')}</p></div><button class="btn primary small" onclick="previewMaterial('${item.id}')">Visualizar</button></div>`).join('')||'<div class="empty">No hay material disponible.</div>'}</div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button></div></div>`);}

function renderIndications() {
  setTop('Indicaciones generales','Administración'); const rules=database.indicationRules||[];
  setContent(`<div class="page-head"><div><div class="eyebrow">Aplicación por alcance</div><h1 class="page-title">Configurar indicaciones</h1><p class="subtitle">Define reglas para todos, por Club, por rol o para una persona.</p></div><button class="btn primary" onclick="openIndicationRule()">Nueva indicación</button></div><div class="list">${rules.map(rule=>{const req=flatRequirements.find(item=>item.id===rule.requirementId);return `<div class="list-item"><div class="req-code">${esc(rule.requirementId)}</div><div><h4>${esc(short(req?.official_text||'',90))}</h4><p>${esc(ruleScopeLabel(rule))} · ${esc(short(rule.config?.instructions||'',100))}</p></div><button class="btn ghost small" onclick="openIndicationRule('${rule.id}')">Editar</button></div>`;}).join('')||'<div class="card empty">No hay indicaciones generales. Se usan las instrucciones oficiales de cada requisito.</div>'}</div>`);
}
function ruleScopeLabel(rule){if(rule.scope==='all')return 'Todos';if(rule.scope==='club')return `Club: ${database.clubs.find(item=>item.id===rule.scopeId)?.name||'—'}`;if(rule.scope==='role')return `Rol: ${ROLE_LABELS[rule.scopeId]||rule.scopeId}`;return `Persona: ${userById(rule.scopeId)?.fullName||'—'}`;}
function openIndicationRule(id='') {
  const rule=database.indicationRules.find(item=>item.id===id),scope=rule?.scope||'all';
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>Indicación general</h3><p>Las reglas más específicas tienen prioridad.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveIndicationRule(event,'${id}')"><div class="modal-body"><label class="field-group">Requisito<select id="rule_req" class="field">${flatRequirements.map(req=>`<option value="${req.id}" ${rule?.requirementId===req.id?'selected':''}>${esc(req.id)} · ${esc(short(req.official_text,70))}</option>`).join('')}</select></label><label class="field-group">Alcance<select id="rule_scope" class="field" onchange="renderRuleTargets()"><option value="all" ${scope==='all'?'selected':''}>Todos</option><option value="club" ${scope==='club'?'selected':''}>Por Club</option><option value="role" ${scope==='role'?'selected':''}>Por rol</option><option value="user" ${scope==='user'?'selected':''}>Por persona</option></select></label><label id="rule_target_wrap" class="field-group hidden">Destino<select id="rule_target" class="field"></select></label><label class="field-group">Modalidad<input id="rule_mode" class="field" value="${esc(rule?.config?.mode||'')}"></label><label class="field-group">Indicaciones<textarea id="rule_instructions" class="field">${esc(rule?.config?.instructions||'')}</textarea></label><label class="field-group">Evaluación<textarea id="rule_evaluation" class="field">${esc(rule?.config?.evaluation||'')}</textarea></label></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary">Guardar</button></div></form></div>`); renderRuleTargets(rule?.scopeId||'');
}
function renderRuleTargets(selected='') { const scope=value('rule_scope'),wrap=document.getElementById('rule_target_wrap'),root=document.getElementById('rule_target'); if(!wrap||!root)return; wrap.classList.toggle('hidden',scope==='all'); const items=scope==='club'?database.clubs.map(x=>[x.id,x.name]):scope==='role'?Object.entries(ROLE_LABELS):database.users.map(x=>[x.id,x.fullName]);root.innerHTML=items.map(([id,label])=>`<option value="${id}" ${id===selected?'selected':''}>${esc(label)}</option>`).join(''); }
async function saveIndicationRule(event,id='') { event.preventDefault();const rule=database.indicationRules.find(item=>item.id===id)||{id:newId('rule'),createdAt:isoNow(),active:true};rule.requirementId=value('rule_req');rule.scope=value('rule_scope');rule.scopeId=rule.scope==='all'?'':value('rule_target');rule.config={mode:value('rule_mode'),instructions:value('rule_instructions'),evaluation:value('rule_evaluation')};if(!id)database.indicationRules.push(rule);await persist(true,{action:'Indicaciones actualizadas',entity:rule.requirementId,detail:ruleScopeLabel(rule)});closeModal();renderIndications();toast('Indicaciones guardadas.');}

function renderRequirements() {
  const aspirant=targetAspirant();
  setTop('Requisitos','Expediente');
  if (!aspirant) return setContent('<div class="card empty">Selecciona o asigna un Aspirante.</div>');
  setContent(`<div class="page-head"><div><div class="eyebrow">Expediente de ${esc(aspirant.fullName)}</div><h1 class="page-title">134 requisitos</h1><p class="subtitle">Guardado local inmediato y sincronización automática al recuperar internet.</p></div>${contextSelector()}</div><div class="toolbar"><input id="req_search" class="field" placeholder="Buscar requisito…" oninput="filterRequirements()"><select id="req_area" class="field" onchange="filterRequirements()"><option value="">Todas las áreas</option>${MASTER.areas.map(area=>`<option value="${area.code}">${esc(area.title)}</option>`).join('')}</select><select id="req_status" class="field" onchange="requirementStatusFilter=this.value;filterRequirements()"><option value="">Todos los estados</option>${Object.entries(STATUS_LABELS).filter(([key])=>key!=='rejected').map(([key,label])=>`<option value="${key}">${label}</option>`).join('')}</select></div><div id="requirements_list" class="list"></div>`);
  const statusRoot=document.getElementById('req_status');
  if(requirementStatusFilter&&statusRoot)statusRoot.value=requirementStatusFilter==='developing'?'':requirementStatusFilter;
  filterRequirements();
}

function filterRequirements() {
  const aspirant=targetAspirant(); if (!aspirant) return;
  const search=normalize(value('req_search')), area=value('req_area'), status=value('req_status'), progress=ensureProgress(aspirant.id),filter=status||requirementStatusFilter;
  const statusMatches=value=>filter==='developing'?['assigned','in_progress','needs_changes'].includes(value):!filter||value===filter;
  const items=flatRequirements.filter(req=>(!area||req.areaCode===area)&&statusMatches(progress.status[req.id])&&(!search||normalize(`${req.id} ${req.official_text} ${req.areaTitle}`).includes(search)));
  document.getElementById('requirements_list').innerHTML=items.map(req=>{const materials=requirementMaterials(req.id);return `<div class="list-item clickable" onclick="openRequirement('${req.id}')">${requirementAreaBadge(req)}<div><h4>${esc(req.official_text)}</h4><div class="chips"><span class="chip">${esc(req.areaTitle)}</span><span class="chip blue">${esc(req.id)}</span>${req.kind==='bible_reading'?'<span class="chip gold">Lectura bíblica</span>':''}<span class="chip blue">${esc(configFor(aspirant.id,req).mode)}</span>${materials.length?`<span class="chip green">${materials.length} material${materials.length===1?'':'es'}</span>`:''}</div></div><div class="req-actions"><span class="status ${progress.status[req.id]}">${STATUS_LABELS[progress.status[req.id]]}</span>${materials.length?`<button class="btn ghost small" onclick="event.stopPropagation();openRequirementMaterials('${req.id}')">Ver material</button>`:''}</div></div>`;}).join('')||'<div class="card empty">No hay resultados.</div>';
}

async function openRequirement(requirementId) {
  const aspirant=targetAspirant(), requirement=flatRequirements.find(item=>item.id===requirementId); if (!aspirant||!requirement) return;
  if (requirement.kind==='bible_reading') return openBibleRequirement(requirementId);
  modalOwnerId=aspirant.id; modalRequirementId=requirementId;
  const submission=submissionFor(aspirant.id,requirementId), config=configFor(aspirant.id,requirement), status=ensureProgress(aspirant.id).status[requirementId];
  const displayedDate=submission.completionDate||localDate();
  const readonly=!['aspirant','pastor_aspirant'].includes(currentUser().role)||currentUser().id!==aspirant.id||['in_review','approved'].includes(status);
  const certificateOnly=false; // SGL 5.0.77: certificado/constancia es evidencia opcional; el requisito usa el flujo normal de informe.
  const options=requirement.selection?.options||requirement.subitems||[];
  const club=clubFor(aspirant),support=requirementMaterials(requirementId);
  const editorHtml=reportHtmlFromSubmission(submission);
  const officialLogo=await loadMandatoryLogo();
  showModal(`<div class="modal"><div class="modal-head"><div><h3>${esc(requirement.areaTitle)}</h3><p>${esc(requirement.official_text)} · ${STATUS_LABELS[status]}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="modal-grid"><section><div class="document-header"><img src="${officialLogo}" alt="Logo oficial"><div class="doc-center">${reportMissionName({aspirant})?`<div class="sgl5053-report-mission">${esc(reportMissionName({aspirant}))}</div>`:''}<strong>INFORME CLASE AGRUPADA</strong><div>${esc(aspirant.fullName)}${club?.logo?'':` · ${esc(club?.name||'Ministerio de Conquistadores')}`}</div></div>${club?.logo?`<img src="${club.logo}" alt="Logo del Club">`:'<span></span>'}</div>${options.length?`<div class="selection-options">${options.map(option=>`<label class="choice"><input type="checkbox" name="req_choice" value="${esc(option)}" ${submission.selections.includes(option)?'checked':''} ${readonly?'disabled':''}> ${esc(option)}</label>`).join('')}</div>`:''}<label class="field-group requirement-general requirement-fill-date-field">Fecha de realización del requisito<input id="req_date" class="field" type="date" value="${esc(displayedDate)}" ${readonly?'disabled':''}></label>${certificateOnly?'<div class="notice">El certificado o constancia es evidencia opcional. Este requisito puede completarse y presentarse mediante el informe normal; si tienes una imagen del certificado, puedes adjuntarla.</div>':`<div class="section-label">Informe</div>${readonly?'':`<div class="rich-toolbar"><button type="button" class="btn ghost small" title="Negrita" onclick="formatReportText('bold')"><b>N</b></button><button type="button" class="btn ghost small" title="Subrayado" onclick="formatReportText('underline')"><u>S</u></button><button type="button" class="btn ghost small" title="Cursiva" onclick="formatReportText('italic')"><i>C</i></button><button type="button" class="btn ghost small" onclick="formatReportText('removeFormat')">Limpiar formato</button><span class="rich-toolbar-sep"></span><button type="button" class="btn ghost small" title="Alinear a la izquierda" onclick="formatReportAlignment('left')">≡←</button><button type="button" class="btn ghost small" title="Centrar" onclick="formatReportAlignment('center')">≡↔</button><button type="button" class="btn ghost small" title="Alinear a la derecha" onclick="formatReportAlignment('right')">→≡</button><button type="button" class="btn ghost small" title="Justificar" onclick="formatReportAlignment('justify')">☰</button><span class="spell-hint">Ortografía en español activada</span></div>`}<div id="req_text" class="sgl5038-writing-page report-area rich-editor ${readonly?'readonly':''}" lang="es" spellcheck="true" contenteditable="${readonly?'false':'true'}" data-placeholder="Escribe aquí el desarrollo del requisito…">${editorHtml}</div>`}<div id="req_files" class="file-grid report-files"></div>${readonly?'':`<label class="upload-zone" style="display:block;margin-top:12px"><strong>${certificateOnly?'Agregar foto del certificado':'Agregar foto o imagen'}</strong><span>Se conserva alta resolución; el informe autoajusta todas las imágenes sin mostrar sus nombres.</span><input type="file" multiple hidden accept="image/*" onchange="attachFiles(this.files)"></label>`}</section><aside>${support.length?`<button class="btn green block" onclick="openRequirementMaterials('${requirementId}')">Ver material de apoyo (${support.length})</button>`:''}<div class="guide-card" style="margin-top:${support.length?'10':'0'}px"><strong>Indicaciones</strong><p>${esc(config.instructions||'Desarrolla el requisito con claridad.')}</p></div><div class="guide-card"><strong>Modalidad</strong><p>${esc(config.mode)}</p></div><div class="guide-card"><strong>Evaluación</strong><p>${esc(config.evaluation)}</p></div><label class="field-group">Orientación del documento<select id="req_orientation" class="field"><option value="auto" ${submission.orientation!=='portrait'&&submission.orientation!=='landscape'?'selected':''}>Automático</option><option value="portrait" ${submission.orientation==='portrait'?'selected':''}>Vertical</option><option value="landscape" ${submission.orientation==='landscape'?'selected':''}>Horizontal</option></select></label><label class="field-group">Letra del informe<select id="req_font" class="field"><option ${aspirant.reportFont!=='Times New Roman'?'selected':''}>Arial</option><option ${aspirant.reportFont==='Times New Roman'?'selected':''}>Times New Roman</option></select></label><button class="btn ghost block" onclick="previewRequirementPdf('${requirementId}')">Visualizar PDF</button><button class="btn primary block" style="margin-top:8px" onclick="exportPdf('${requirementId}')">Exportar PDF</button><button class="btn ghost block" style="margin-top:8px" onclick="printRequirement('${requirementId}')">Imprimir</button>${submission.feedback?`<div class="notice warn" style="margin-top:10px"><b>Observación del revisor</b><br>${esc(submission.feedback)}</div>`:''}</aside></div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button>${!readonly?`<button class="btn ghost" onclick="saveRequirement(false)">Guardar avance</button>`:''}${!['aspirant','pastor_aspirant'].includes(currentUser().role)&&status==='in_review'?`<button class="btn warn" onclick="reviewRequirement('needs_changes')">Solicitar corrección</button><button class="btn green" onclick="reviewRequirement('approved')">Marcar revisado</button>`:''}</div></div>`);
  await renderRequirementFiles();
  if(!readonly)enableRequirementAutosave();
}

async function attachFiles(fileList) {
  const aspirant=userById(modalOwnerId), submission=submissionFor(modalOwnerId,modalRequirementId);
  for (const file of [...fileList]) {
    if(!String(file.type||'').startsWith('image/')){toast('En Informes/Requisitos solo se aceptan fotos o imágenes.');continue;}
    const dataUrl=await imageToDataUrl(file,3*1024*1024,3200);
    const attachment={id:newId('file'),ownerId:aspirant.id,requirementId:modalRequirementId,name:file.name,type:file.type||'image/jpeg',size:dataUrlBytes(dataUrl),kind:'photo',dataUrl,reportPlacement:'auto',reportScale:70,createdAt:isoNow()};
    await filePut(attachment); submission.attachmentIds.push(attachment.id);
  }
  submission.updatedAt=isoNow(); await persist(true,{action:'Archivos adjuntados',entity:modalRequirementId,detail:`${fileList.length} archivo(s)`}); await renderRequirementFiles(); toast('Archivos guardados en este dispositivo.');
}

async function renderRequirementFiles() {
  const root=document.getElementById('req_files'); if (!root) return;
  const submission=submissionFor(modalOwnerId,modalRequirementId), files=await allFiles(), allowed=files.filter(file=>submission.attachmentIds.includes(file.id)&&(file.kind==='photo'||String(file.type||'').startsWith('image/')));
  root.innerHTML=allowed.map(file=>`<div class="file-card"><img src="${file.dataUrl}" alt="${esc(file.name)}" onclick="openAttachment('${file.id}')"><button onclick="event.stopPropagation();removeAttachment('${file.id}')">×</button><div>${esc(short(file.name,28))}</div><select class="field compact" onchange="updateAttachmentLayout('${file.id}','reportPlacement',this.value)"><option value="auto" ${file.reportPlacement==='auto'?'selected':''}>Autoajustar</option><option value="page1" ${file.reportPlacement==='page1'?'selected':''}>Página 1</option><option value="page2" ${file.reportPlacement==='page2'?'selected':''}>Página 2</option></select><label class="mini-range">Tamaño ${file.reportScale||70}%<input type="range" min="1" max="100" step="1" value="${file.reportScale||70}" onchange="updateAttachmentLayout('${file.id}','reportScale',this.value)"></label></div>`).join('');
}

async function updateAttachmentLayout(id,key,next){const file=(await allFiles()).find(item=>item.id===id);if(!file)return;file[key]=key==='reportScale'?Number(next):next;await filePut(file);await persistDraftLocal();await renderRequirementFiles();}

function closeImageViewer(fromHistory=false){
  const root=document.getElementById('sgl_image_viewer');
  if(root)root.remove();
  sglImageViewer=null;
  document.documentElement.classList.remove('sgl-photo-viewer-open');
  if(!fromHistory&&history.state?.sgl&&history.state.kind==='photo')sglDiscardTransientHistory();
}
function sglImageViewerTransform(){
  if(!sglImageViewer)return;
  const image=document.getElementById('sgl_viewer_image'),label=document.getElementById('sgl_viewer_zoom');
  if(!image)return;
  image.style.transform=`translate3d(${sglImageViewer.x}px,${sglImageViewer.y}px,0) scale(${sglImageViewer.scale})`;
  if(label)label.textContent=`${Math.round(sglImageViewer.scale*100)}%`;
}
function sglImageViewerReset(){
  if(!sglImageViewer)return;
  sglImageViewer.scale=1;sglImageViewer.x=0;sglImageViewer.y=0;sglImageViewerTransform();
}
function sglImageViewerZoom(delta){
  if(!sglImageViewer)return;
  sglImageViewer.scale=Math.max(.35,Math.min(8,sglImageViewer.scale+delta));
  sglImageViewerTransform();
}
function sglImageViewerZoomAt(nextScale,clientX,clientY){
  if(!sglImageViewer)return;
  const stage=document.getElementById('sgl_viewer_stage');
  if(!stage)return;
  const rect=stage.getBoundingClientRect(),old=sglImageViewer.scale,next=Math.max(.35,Math.min(8,nextScale));
  if(Math.abs(next-old)<.0001)return;
  const ox=clientX-(rect.left+rect.width/2)-sglImageViewer.x;
  const oy=clientY-(rect.top+rect.height/2)-sglImageViewer.y;
  const ratio=next/old;
  sglImageViewer.x-=ox*(ratio-1);
  sglImageViewer.y-=oy*(ratio-1);
  sglImageViewer.scale=next;
  sglImageViewerTransform();
}
function openImageViewer(source,name='Fotografía'){
  if(!source)return;
  if(document.getElementById('sgl_image_viewer'))closeImageViewer(true);
  sglPushTransientHistory('photo');
  const root=document.createElement('div');root.id='sgl_image_viewer';root.className='sgl-image-viewer';
  root.innerHTML=`<div class="sgl-viewer-toolbar"><div class="sgl-viewer-title"><strong>EVIDENCIA FOTOGRÁFICA</strong><small>${esc(name||'Fotografía')}</small></div><div class="sgl-viewer-actions"><button type="button" onclick="sglImageViewerZoom(-.25)" aria-label="Alejar" title="Alejar">−</button><span id="sgl_viewer_zoom">100%</span><button type="button" onclick="sglImageViewerZoom(.25)" aria-label="Acercar" title="Acercar">+</button><button type="button" class="sgl-viewer-fit" onclick="sglImageViewerReset()">Ajustar</button><button type="button" class="sgl-viewer-close" onclick="closeImageViewer()" aria-label="Cerrar" title="Cerrar">×</button></div></div><div class="sgl-viewer-stage" id="sgl_viewer_stage"><img id="sgl_viewer_image" draggable="false" src="${source}" alt="${esc(name||'Fotografía')}"><div class="sgl-viewer-hint">Mueve la foto con mouse o dedo · rueda o pellizca para zoom · doble toque para ajustar</div></div>`;
  document.body.appendChild(root);document.documentElement.classList.add('sgl-photo-viewer-open');
  const stage=document.getElementById('sgl_viewer_stage'),image=document.getElementById('sgl_viewer_image');
  const pointers=new Map();
  sglImageViewer={scale:1,x:0,y:0,lastDistance:0,lastTap:0,dragging:false};
  image?.addEventListener('load',()=>sglImageViewerReset(),{once:true});
  stage.addEventListener('wheel',event=>{event.preventDefault();sglImageViewerZoomAt(sglImageViewer.scale*(event.deltaY<0?1.16:.86),event.clientX,event.clientY);},{passive:false});
  stage.addEventListener('dblclick',event=>{event.preventDefault();sglImageViewerReset();});
  const pointerDown=event=>{
    stage.setPointerCapture?.(event.pointerId);pointers.set(event.pointerId,{x:event.clientX,y:event.clientY});sglImageViewer.dragging=true;
    if(pointers.size===2){const p=[...pointers.values()];sglImageViewer.lastDistance=Math.hypot(p[0].x-p[1].x,p[0].y-p[1].y);}
    event.preventDefault();
  };
  const pointerMove=event=>{
    if(!pointers.has(event.pointerId)||!sglImageViewer)return;
    const previous=pointers.get(event.pointerId);pointers.set(event.pointerId,{x:event.clientX,y:event.clientY});
    if(pointers.size===1){sglImageViewer.x+=event.clientX-previous.x;sglImageViewer.y+=event.clientY-previous.y;sglImageViewerTransform();}
    else if(pointers.size>=2){
      const p=[...pointers.values()].slice(0,2),distance=Math.hypot(p[0].x-p[1].x,p[0].y-p[1].y),midX=(p[0].x+p[1].x)/2,midY=(p[0].y+p[1].y)/2;
      if(sglImageViewer.lastDistance>0)sglImageViewerZoomAt(sglImageViewer.scale*(distance/sglImageViewer.lastDistance),midX,midY);
      sglImageViewer.lastDistance=distance;
    }
    event.preventDefault();
  };
  const pointerRelease=event=>{
    if(!pointers.has(event.pointerId))return;
    pointers.delete(event.pointerId);sglImageViewer.lastDistance=0;sglImageViewer.dragging=pointers.size>0;
    const now=Date.now();if(event.pointerType==='touch'&&now-sglImageViewer.lastTap<330)sglImageViewerReset();sglImageViewer.lastTap=now;
  };
  if(window.PointerEvent){
    stage.addEventListener('pointerdown',pointerDown);stage.addEventListener('pointermove',pointerMove);stage.addEventListener('pointerup',pointerRelease);stage.addEventListener('pointercancel',pointerRelease);
  }else{
    let mouseDown=false,lastMouse=null,touchDistance=0;
    stage.addEventListener('mousedown',event=>{mouseDown=true;lastMouse={x:event.clientX,y:event.clientY};event.preventDefault();});
    window.addEventListener('mousemove',event=>{if(!mouseDown||!lastMouse||!sglImageViewer)return;sglImageViewer.x+=event.clientX-lastMouse.x;sglImageViewer.y+=event.clientY-lastMouse.y;lastMouse={x:event.clientX,y:event.clientY};sglImageViewerTransform();});
    window.addEventListener('mouseup',()=>{mouseDown=false;lastMouse=null;},{once:false});
    stage.addEventListener('touchstart',event=>{if(event.touches.length===1){lastMouse={x:event.touches[0].clientX,y:event.touches[0].clientY};}else if(event.touches.length>=2){const a=event.touches[0],b=event.touches[1];touchDistance=Math.hypot(a.clientX-b.clientX,a.clientY-b.clientY);}event.preventDefault();},{passive:false});
    stage.addEventListener('touchmove',event=>{if(!sglImageViewer)return;if(event.touches.length===1&&lastMouse){const t=event.touches[0];sglImageViewer.x+=t.clientX-lastMouse.x;sglImageViewer.y+=t.clientY-lastMouse.y;lastMouse={x:t.clientX,y:t.clientY};sglImageViewerTransform();}else if(event.touches.length>=2){const a=event.touches[0],b=event.touches[1],distance=Math.hypot(a.clientX-b.clientX,a.clientY-b.clientY),midX=(a.clientX+b.clientX)/2,midY=(a.clientY+b.clientY)/2;if(touchDistance>0)sglImageViewerZoomAt(sglImageViewer.scale*(distance/touchDistance),midX,midY);touchDistance=distance;}event.preventDefault();},{passive:false});
    stage.addEventListener('touchend',event=>{const now=Date.now();if(event.touches.length===0&&now-sglImageViewer.lastTap<330)sglImageViewerReset();sglImageViewer.lastTap=now;lastMouse=null;touchDistance=0;},{passive:false});
  }
  sglImageViewerTransform();
}
async function openAttachment(id) {
  const file=(await allFiles()).find(item=>item.id===id);if(!file)return;
  const isImage=file.kind==='photo'||String(file.type||'').startsWith('image/')||String(file.dataUrl||'').startsWith('data:image/');
  if(isImage){openImageViewer(file.dataUrl,file.name||'Fotografía');return;}
  const popup=window.open(file.dataUrl,'_blank');if(!popup)toast('Permite ventanas emergentes para abrir el archivo.');
}

async function removeAttachment(id) {
  const submission=submissionFor(modalOwnerId,modalRequirementId); submission.attachmentIds=submission.attachmentIds.filter(item=>item!==id); await fileDelete(id); await persistDraftLocal(); await renderRequirementFiles();
}

async function saveRequirement(send=false) {
  const submission=submissionFor(modalOwnerId,modalRequirementId), progress=ensureProgress(modalOwnerId), requirement=flatRequirements.find(item=>item.id===modalRequirementId), config=configFor(modalOwnerId,requirement), certificateOnly=false; // SGL 5.0.77: certificado/constancia nunca bloquea el informe normal.
  const editor=readRichEditor();
  if(!editor.text)return toast('Escribe el informe antes de guardar. Las fotografías y selecciones complementan el texto, pero no lo reemplazan.');
  submission.text=editor.text; submission.richText=editor.html; submission.completionDate=value('req_date'); submission.orientation=value('req_orientation')||submission.orientation||'auto'; submission.selections=[...document.querySelectorAll('[name=req_choice]:checked')].map(input=>input.value); submission.updatedAt=isoNow();
  const aspirant=userById(modalOwnerId);if(aspirant&&value('req_font'))aspirant.reportFont=value('req_font');
  if (send && !submission.completionDate) return toast('Completa la fecha antes de enviar.');
  if (send && !submission.text) return toast('Completa el informe antes de enviar.');
  progress.status[modalRequirementId]=send?'in_review':'in_progress'; progress.updatedAt=isoNow();
  if(send)for(const task of database.tasks.filter(item=>item.assigneeId===modalOwnerId&&item.requirementId===modalRequirementId&&!item.presentedAt)){task.presentedAt=isoNow();task.updatedAt=isoNow();}
  await persist(true,{action:send?'Requisito enviado':'Avance guardado',entity:modalRequirementId,detail:userById(modalOwnerId)?.fullName||''}); closeModal(); render(); toast(send?'Enviado a revisión.':'Avance guardado.');
  localStorage.removeItem('musica_emergency_draft');
}

function captureRequirementDraft() {
  if(!modalOwnerId||!modalRequirementId||!document.getElementById('req_date'))return null;
  const editor=readRichEditor();return {type:'requirement',ownerId:modalOwnerId,requirementId:modalRequirementId,text:editor.text,richText:editor.html,completionDate:value('req_date')||localDate(),orientation:value('req_orientation')||'auto',reportFontSize:Math.max(10,Math.min(14,Number(document.getElementById('sgl5047_report_font_size')?.value||submissionFor(modalOwnerId,modalRequirementId)?.reportFontSize||12)||12)),selections:[...document.querySelectorAll('[name=req_choice]:checked')].map(input=>input.value),savedAt:isoNow()};
}
function saveEmergencyDraft(){const draft=captureRequirementDraft();if(draft)try{localStorage.setItem('musica_emergency_draft',JSON.stringify(draft));}catch{}}
function enableRequirementAutosave(){const root=document.querySelector('.modal-bg');if(!root)return;const save=()=>{saveEmergencyDraft();clearTimeout(draftTimer);draftTimer=setTimeout(flushRequirementDraft,900);};root.addEventListener('input',save);root.addEventListener('change',save);const emergency=JSON.parse(localStorage.getItem('musica_emergency_draft')||'null');if(emergency?.ownerId===modalOwnerId&&emergency?.requirementId===modalRequirementId&&emergency.savedAt>submissionFor(modalOwnerId,modalRequirementId).updatedAt){const text=document.getElementById('req_text');if(text)text.innerHTML=emergency.richText||plainTextToRichHtml(emergency.text||'');const date=document.getElementById('req_date');if(date)date.value=emergency.completionDate||localDate();toast('Se recuperó el borrador del cierre anterior.');}}
async function flushRequirementDraft(){saveEmergencyDraft();}

function formatReportText(command){const editor=document.getElementById('req_text');if(!editor)return;editor.focus();document.execCommand(command,false,null);}
function formatReportAlignment(value){const editor=document.getElementById('req_text');if(!editor)return;const align=['left','center','right','justify'].includes(value)?value:'justify',command={left:'justifyLeft',center:'justifyCenter',right:'justifyRight',justify:'justifyFull'}[align];editor.focus();document.execCommand(command,false,null);for(const node of editor.querySelectorAll('[align],p,div')){const raw=(node.getAttribute('align')||node.style?.textAlign||'').toLowerCase();if(['left','center','right','justify'].includes(raw))node.style.textAlign=raw;node.removeAttribute('align');}}
function plainTextToRichHtml(text){return String(text||'').split(/\n{2,}/).filter(Boolean).map(paragraph=>`<p>${esc(paragraph).replace(/\n/g,'<br>')}</p>`).join('');}

function reportPlainTextFromHtml(html){
  const container=document.createElement('div');container.innerHTML=sanitizeRichText(html||'');
  container.querySelectorAll('[data-sgl-editor-ui],figure[data-sgl-photo-id]').forEach(node=>node.remove());
  container.querySelectorAll('br').forEach(node=>node.replaceWith(document.createTextNode('\n')));
  container.querySelectorAll('p,div,li,h1,h2,h3,h4,h5,h6,tr,blockquote').forEach(node=>node.append(document.createTextNode('\n')));
  container.querySelectorAll('td,th').forEach(node=>node.append(document.createTextNode(' ')));
  return String(container.textContent||'').replace(/\u200b/g,'').replace(/ /g,' ').replace(/[ \t]+\n/g,'\n').replace(/\n{3,}/g,'\n\n').trim();
}
function reportNormalizedText(text){return String(text||'').replace(/ /g,' ').replace(/\s+/g,' ').trim();}
function reportHtmlFromSubmission(submission={}){
  const plain=String(submission.text||'').replace(/ /g,' ').trim(),rich=sanitizeRichText(submission.richText||''),richText=reportPlainTextFromHtml(rich),nPlain=reportNormalizedText(plain),nRich=reportNormalizedText(richText),richHasPhotos=/data-sgl-photo-id=["'][^"']+["']/i.test(rich);
  // 5.0.62: un informe puede contener una fotografia aunque no tenga texto. Nunca
  // degradar ese HTML a cadena vacia o a texto plano, porque eso elimina la figura
  // y al reabrir obliga a reconstruirla con la colocacion automatica centrada.
  if(!nPlain&&!nRich)return richHasPhotos||/<(?:p|div|br)\b/i.test(rich)?rich:'';
  if(nPlain&&nRich&&nPlain.replace(/\s/g,'')===nRich.replace(/\s/g,''))return rich;
  if(nPlain&&(!nRich||nRich.length<nPlain.length*.85||!nPlain.includes(nRich.slice(0,Math.min(24,nRich.length)))))return richHasPhotos?rich:plainTextToRichHtml(plain);
  return rich||plainTextToRichHtml(plain);
}
function reportSubmissionCacheKey(submission={}){return [submission.updatedAt||'',submission.completionDate||'',Number(submission.reportFontSize||12)||12,reportNormalizedText(submission.text||''),sanitizeRichText(submission.richText||''),JSON.stringify(submission.selections||[]),JSON.stringify(submission.reportWatermark||null),JSON.stringify(submission.reportPhotoBoxes||{})];}
function liveRequirementSubmission(baseSubmission={},ownerId=modalOwnerId,requirementId=modalRequirementId){
  const snapshot=structuredClone(baseSubmission||{}),editor=document.getElementById('req_text');
  if(editor&&ownerId===modalOwnerId&&requirementId===modalRequirementId){
    const current=readRichEditor();
    if(current.text||current.html){snapshot.text=current.text;snapshot.richText=current.html;}
    const date=value('req_date');if(date)snapshot.completionDate=date;
    const orientation=value('req_orientation');if(orientation)snapshot.orientation=orientation;
    const fontSize=document.getElementById('sgl5047_report_font_size')?.value;if(fontSize)snapshot.reportFontSize=Math.max(10,Math.min(14,Number(fontSize)||12));
    snapshot.selections=[...document.querySelectorAll('[name=req_choice]:checked')].map(input=>input.value);
  }
  snapshot.richText=reportHtmlFromSubmission(snapshot);
  return snapshot;
}
function normalizeWordClipboardHtml(html=''){
  const template=document.createElement('template');template.innerHTML=String(html||'');
  // 5.0.62: Word suele guardar la tipografia real en reglas de clases MsoNormal,
  // MsoListParagraph, etc. Se copian solo propiedades textuales permitidas a style
  // inline antes de retirar las hojas de estilo del portapapeles.
  const safeProps=new Set(['font-family','font-size','font-weight','font-style','text-decoration','text-decoration-line','text-align','text-indent','line-height','letter-spacing','vertical-align','margin','margin-top','margin-right','margin-bottom','margin-left','padding-left','list-style-type']);
  const normalizeProp=prop=>{
    const key=String(prop||'').trim().toLowerCase();
    if(key==='mso-bidi-font-weight')return 'font-weight';
    if(key==='mso-bidi-font-style')return 'font-style';
    return safeProps.has(key)?key:'';
  };
  const applyDeclarations=(node,raw='')=>{
    if(!node?.style)return;
    for(const part of String(raw||'').split(';')){
      const cut=part.indexOf(':');if(cut<1)continue;
      const prop=normalizeProp(part.slice(0,cut));if(!prop)continue;
      let value=part.slice(cut+1).trim();if(!value||/[{}<>]/.test(value)||/(?:url|expression|var)\s*\(/i.test(value))continue;
      const important=/\s*!important\s*$/i.test(value);value=value.replace(/\s*!important\s*$/i,'').trim();
      try{node.style.setProperty(prop,value,important?'important':'');}catch{}
    }
  };
  for(const style of [...template.content.querySelectorAll('style')]){
    const css=String(style.textContent||'');
    for(const match of css.matchAll(/([^{}]+)\{([^{}]*)\}/g)){
      const declarations=match[2]||'';
      for(const rawSelector of String(match[1]||'').split(',')){
        const selector=rawSelector.trim();
        // Word usa selectores simples como p.MsoNormal y p.MsoListParagraphCxSpFirst.
        if(!/^(?:[a-z][a-z0-9-]*)?(?:\.[A-Za-z_][A-Za-z0-9_-]*)+$/i.test(selector))continue;
        let nodes=[];try{nodes=[...template.content.querySelectorAll(selector)];}catch{}
        nodes.forEach(node=>applyDeclarations(node,declarations));
      }
    }
  }
  // Las hojas de estilo y metadatos de Word ya fueron interpretados. No deben
  // sobrevivir al saneado ni convertirse accidentalmente en texto visible.
  template.content.querySelectorAll('style,script,meta,link,xml').forEach(node=>node.remove());
  // Compatibilidad con fragmentos Word antiguos que todavía usan <font>.
  for(const font of [...template.content.querySelectorAll('font')]){
    const span=document.createElement('span'),face=String(font.getAttribute('face')||'').trim(),size=Number(font.getAttribute('size')||0);
    if(face)span.style.fontFamily=face;
    const htmlSizePt={1:8,2:10,3:12,4:14,5:18,6:24,7:36}[size];if(htmlSizePt)span.style.fontSize=`${htmlSizePt}pt`;
    while(font.firstChild)span.appendChild(font.firstChild);font.replaceWith(span);
  }
  return template.innerHTML;
}
window.normalizeWordClipboardHtml=normalizeWordClipboardHtml;

function sgl532Color(value){
  const probe=document.createElement('span');probe.style.color=String(value||'');
  const raw=probe.style.color,hex=raw.match(/^#([0-9a-f]{6})$/i),rgb=raw.match(/^rgb\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)$/i);
  if(hex)return hex[1].toUpperCase();
  if(rgb)return rgb.slice(1).map(n=>Math.max(0,Math.min(255,Number(n))).toString(16).padStart(2,'0')).join('').toUpperCase();
  return '';
}
function sgl532PdfColor(value){const hex=sgl532Color(value);return hex?[0,2,4].map(i=>(parseInt(hex.slice(i,i+2),16)/255).toFixed(4)).join(' '):'';}
function sanitizeRichText(html){
  const template=document.createElement('template');template.innerHTML=normalizeWordClipboardHtml(html||'');
  // 5.0.62: conservar el formato textual seguro que Word coloca en los
  // párrafos seleccionados sin cambiar la configuración normal del editor.
  // La ruta de fotografías/figuras SGL se mantiene sin cambios funcionales.
  const allowed=new Set(['P','DIV','BR','B','STRONG','U','I','EM','SPAN','S','STRIKE','SUP','SUB','UL','OL','LI','FIGURE','IMG','TABLE','THEAD','TBODY','TFOOT','TR','TD','TH','COLGROUP','COL','CAPTION']),alignments=new Set(['left','center','right','justify']);
  const safeFontFamily=value=>{
    const raw=String(value||'').trim();
    if(!raw||raw.length>160||/[;{}<>]/.test(raw)||/(?:url|expression|var)\s*\(/i.test(raw))return '';
    return /^[A-Za-z0-9À-ÿ _'",.\-]+$/.test(raw)?raw:'';
  };
  const safeFontSize=value=>{
    const raw=String(value||'').trim().toLowerCase(),match=raw.match(/^(\d+(?:\.\d+)?)(pt|px|em|rem|%)$/);if(!match)return '';
    const n=Number(match[1]),unit=match[2],limits={pt:[6,72],px:[8,96],em:[.5,6],rem:[.5,6],'%':[50,600]}[unit];
    return limits&&n>=limits[0]&&n<=limits[1]?`${n}${unit}`:'';
  };
  const safeLength=(value,{negative=false,percent=true}={})=>{
    const raw=String(value||'').trim().toLowerCase(),match=raw.match(/^(-?\d+(?:\.\d+)?)(pt|px|em|rem|%|cm|mm|in)$/);if(!match)return '';
    const n=Number(match[1]),unit=match[2];if(!negative&&n<0)return '';if(!percent&&unit==='%')return '';
    const limit={pt:360,px:480,em:30,rem:30,'%':200,cm:13,mm:130,in:5}[unit];if(!Number.isFinite(n)||Math.abs(n)>limit)return '';
    return `${n}${unit}`;
  };
  const safeLineHeight=value=>{
    const raw=String(value||'').trim().toLowerCase();if(raw==='normal')return 'normal';
    if(/^\d+(?:\.\d+)?$/.test(raw)){const n=Number(raw);return n>=.7&&n<=4?String(n):'';}
    const length=safeLength(raw,{negative:false,percent:true});if(!length)return '';
    if(length.endsWith('%')){const n=Number.parseFloat(length);return n>=70&&n<=400?length:'';}
    return length;
  };
  const clean=node=>{
    for(const child of [...node.childNodes]){
      if(child.nodeType!==1)continue;
      if(child.hasAttribute?.('data-sgl-editor-ui')||String(child.getAttribute?.('data-sgl-photo-id')||'').startsWith('pending:')){child.remove();continue;}
      clean(child);
      if(!allowed.has(child.tagName)){child.replaceWith(...child.childNodes);continue;}
      const rawStyle=String(child.getAttribute('style')||''),rawAlign=String(child.style?.textAlign||child.getAttribute('align')||'').toLowerCase(),safeAlign=alignments.has(rawAlign)?rawAlign:'';
      const rawWeight=String(child.style?.fontWeight||'').toLowerCase(),weightNumber=parseInt(rawWeight,10),styleBold=rawWeight==='bold'||rawWeight==='bolder'||Number.isFinite(weightNumber)&&weightNumber>=600||/mso-(?:bidi-)?font-weight\s*:\s*bold/i.test(rawStyle);
      const rawFontStyle=String(child.style?.fontStyle||'').toLowerCase(),styleItalic=rawFontStyle==='italic'||rawFontStyle==='oblique'||/mso-(?:bidi-)?font-style\s*:\s*italic/i.test(rawStyle);
      const rawDecoration=`${child.style?.textDecoration||''} ${child.style?.textDecorationLine||''}`.toLowerCase(),styleUnderline=rawDecoration.includes('underline')||/text-decoration(?:-line)?\s*:[^;]*underline/i.test(rawStyle),styleStrike=rawDecoration.includes('line-through');
      const textNode=child.tagName!=='FIGURE'&&child.tagName!=='IMG',blockText=['P','DIV','LI'].includes(child.tagName),fontFamily=textNode?safeFontFamily(child.style?.fontFamily):'',fontSize=textNode?safeFontSize(child.style?.fontSize):'',lineHeight=blockText?safeLineHeight(child.style?.lineHeight):'',marginTop=blockText?safeLength(child.style?.marginTop):'',marginRight=blockText?safeLength(child.style?.marginRight):'',marginBottom=blockText?safeLength(child.style?.marginBottom):'',marginLeft=blockText?safeLength(child.style?.marginLeft):'',textIndent=blockText?safeLength(child.style?.textIndent,{negative:true}):'',letterSpacing=textNode?safeLength(child.style?.letterSpacing,{negative:true,percent:false}):'',verticalAlign=textNode&&['baseline','sub','super'].includes(String(child.style?.verticalAlign||'').toLowerCase())?String(child.style.verticalAlign).toLowerCase():'',listStyle=['disc','circle','square','decimal','lower-alpha','upper-alpha','lower-roman','upper-roman'].includes(String(child.style?.listStyleType||'').toLowerCase())?String(child.style.listStyleType).toLowerCase():'';
      const isTable=child.tagName==='TABLE',isCell=['TD','TH'].includes(child.tagName);
      const colspan=isCell?Math.max(1,Math.min(100,parseInt(child.getAttribute('colspan'))||1)):1;
      const rowspan=isCell?Math.max(1,Math.min(1000,parseInt(child.getAttribute('rowspan'))||1)):1;
      const cellAlign=['top','middle','bottom'].includes(child.style.verticalAlign)?child.style.verticalAlign:'top';
      const tableWidth=/^(?:\d+(?:\.\d+)?)(?:%|px|pt)$/.test(child.style.width)?child.style.width:'';
      // A row height is a minimum; growing text must always remain visible.
      const rowHeight=child.tagName==='TR'&&/^(?:\d+(?:\.\d+)?)(?:px|pt)$/.test(child.style.height)?Math.max(1,Math.min(600,parseFloat(child.style.height)*(child.style.height.endsWith('pt')?4/3:1)))+'px':'';
      const pageBreak=child.tagName==='P'&&child.getAttribute('data-sgl-page-break')==='1';
      const photoLayout=['inline','square','tight','through','topbottom','behind','front'].includes(child.getAttribute('data-sgl-layout'))?child.getAttribute('data-sgl-layout'):'';
      const photoAngle=Number(child.getAttribute('data-sgl-angle'))||0,photoAnchor=child.getAttribute('data-sgl-anchor')==='text'?'text':'page';
      const isFigure=child.tagName==='FIGURE',photoId=isFigure?String(child.getAttribute('data-sgl-photo-id')||'').replace(/[^A-Za-z0-9_.:-]/g,''):'';
      const scale=isFigure?Math.max(1,Math.min(100,Number(child.getAttribute('data-sgl-scale')||70)||70)):70;
      const photoAlign=isFigure&&['left','center','right'].includes(String(child.getAttribute('data-sgl-align')||''))?String(child.getAttribute('data-sgl-align')):'center';
      const rawXText=isFigure?String(child.getAttribute('data-sgl-x')??'').trim():'',rawX=rawXText===''?NaN:Number(rawXText),photoX=isFigure?(Number.isFinite(rawX)?Math.max(0,Math.min(100,rawX)):(photoAlign==='left'?0:photoAlign==='right'?100:50)):50;
      const free=isFigure&&String(child.getAttribute('data-sgl-free')||'')==='1';
      const left=isFigure?Math.max(0,Math.min(12000,Number(child.getAttribute('data-sgl-left')||0)||0)):0;
      const top=isFigure?Math.max(0,Math.min(60000,Number(child.getAttribute('data-sgl-top')||0)||0)):0;
      const z=isFigure?Math.max(1,Math.min(9999,Math.round(Number(child.getAttribute('data-sgl-z')||1)||1))):1;
      const rawPage=isFigure?String(child.getAttribute('data-sgl-page')||'').trim():'',photoPage=rawPage?Math.max(1,Math.min(999,Math.round(Number(rawPage)||1))):null;
      const rawHeight=isFigure?Number(child.getAttribute('data-sgl-height')||0):0,photoHeight=Number.isFinite(rawHeight)?Math.max(0,Math.min(12000,rawHeight)):0;
      const rawAspect=isFigure?Number(child.getAttribute('data-sgl-aspect')||0):0,photoAspect=Number.isFinite(rawAspect)?Math.max(0,Math.min(20,rawAspect)):0;
      const alt=child.tagName==='IMG'?String(child.getAttribute('alt')||'Evidencia fotográfica').slice(0,160):'';
      const color=sgl532Color(child.style.color),background=sgl532Color(child.style.backgroundColor);
      for(const attribute of [...child.attributes])child.removeAttribute(attribute.name);
      if(pageBreak)child.setAttribute('data-sgl-page-break','1');
      const safeStyles=[];

      if(isTable)safeStyles.push('border-collapse:collapse','width:100%','max-width:100%','table-layout:fixed');
      if(isCell){
        if(colspan>1)child.setAttribute('colspan',String(colspan));
        if(rowspan>1)child.setAttribute('rowspan',String(rowspan));
        safeStyles.push('border:1px solid #555','padding:5px','vertical-align:'+cellAlign,'overflow-wrap:anywhere');
        if(tableWidth)safeStyles.push('width:'+tableWidth);
        if(safeAlign)safeStyles.push('text-align:'+safeAlign);
      }
      if(child.tagName==='COL'&&tableWidth)safeStyles.push('width:'+tableWidth);
      if(rowHeight)safeStyles.push('height:'+rowHeight);
      if(blockText&&safeAlign)safeStyles.push(`text-align:${safeAlign}`);
      if(textNode){
        if(color)safeStyles.push(`color:#${color}!important`);
        if(background)safeStyles.push(`background-color:#${background}`);
        if(rawWeight==='normal'||rawWeight==='400')safeStyles.push('font-weight:400');
        if(rawFontStyle==='normal')safeStyles.push('font-style:normal');
        if(rawDecoration.trim()==='none')safeStyles.push('text-decoration:none');
        if(fontFamily)safeStyles.push(`font-family:${fontFamily}`);
        if(fontSize)safeStyles.push(`font-size:${fontSize}`);
        if(styleBold)safeStyles.push('font-weight:700');
        if(styleItalic)safeStyles.push('font-style:italic');
        if(styleUnderline||styleStrike)safeStyles.push(`text-decoration:${[styleUnderline?'underline':'',styleStrike?'line-through':''].filter(Boolean).join(' ')}`);
        if(letterSpacing)safeStyles.push(`letter-spacing:${letterSpacing}`);
        if(verticalAlign)safeStyles.push(`vertical-align:${verticalAlign}`);
      }
      // Las reglas de la hoja Carta usan !important para su configuración base.
      // Solo un párrafo pegado que trae explícitamente estas medidas desde Word
      // recibe valores inline importantes; un párrafo escrito en SGL sigue
      // heredando exactamente la configuración normal del informe.
      if(['UL','OL'].includes(child.tagName)&&listStyle)safeStyles.push(`list-style-type:${listStyle}`);
      if(blockText){
        if(lineHeight)safeStyles.push(`line-height:${lineHeight}!important`);
        if(marginTop)safeStyles.push(`margin-top:${marginTop}!important`);
        if(marginRight)safeStyles.push(`margin-right:${marginRight}!important`);
        if(marginBottom)safeStyles.push(`margin-bottom:${marginBottom}!important`);
        if(marginLeft)safeStyles.push(`margin-left:${marginLeft}!important`);
        if(textIndent)safeStyles.push(`text-indent:${textIndent}`);
      }
      if(safeStyles.length)child.setAttribute('style',safeStyles.join(';'));
      if(isFigure){
        if(photoLayout)child.setAttribute('data-sgl-layout',photoLayout);
        child.setAttribute('data-sgl-angle',String(((photoAngle%360)+360)%360));child.setAttribute('data-sgl-anchor',photoAnchor);
        if(!photoId){child.remove();continue;}
        child.setAttribute('class','sgl-inline-photo');child.setAttribute('data-sgl-photo-id',photoId);child.setAttribute('data-sgl-scale',String(scale));child.setAttribute('data-sgl-align',photoAlign);child.setAttribute('data-sgl-x',String(Math.round(photoX*10)/10));child.setAttribute('contenteditable','false');if(photoHeight>0)child.setAttribute('data-sgl-height',String(Math.round(photoHeight*10)/10));if(photoAspect>0)child.setAttribute('data-sgl-aspect',String(Math.round(photoAspect*10000)/10000));
        if(free){child.setAttribute('data-sgl-free','1');child.setAttribute('data-sgl-left',String(Math.round(left*10)/10));child.setAttribute('data-sgl-top',String(Math.round(top*10)/10));child.setAttribute('data-sgl-z',String(z));if(photoPage)child.setAttribute('data-sgl-page',String(photoPage));}
      }
      if(child.tagName==='IMG')child.setAttribute('alt',alt||'Evidencia fotográfica');
    }
  };
  clean(template.content);return template.innerHTML;
}
function sgl512InlinePhotoIds(html){const box=document.createElement('div');box.innerHTML=sanitizeRichText(html||'');return new Set([...box.querySelectorAll('figure[data-sgl-photo-id]')].map(node=>String(node.dataset.sglPhotoId||'')).filter(Boolean));}
function sgl514LayoutInlineReportFigures(box){
  let previousRight=null,inPhotoRun=false;
  for(const node of [...box.childNodes]){
    const figure=node.nodeType===1&&node.tagName==='FIGURE'&&node.dataset.sglPhotoId?node:null;
    if(!figure){if(node.nodeType===1&&['P','DIV'].includes(node.tagName))node.style.clear='both';previousRight=null;inPhotoRun=false;continue;}
    const scale=Math.max(1,Math.min(100,Number(figure.dataset.sglScale||70)||70));
    if(figure.dataset.sglFree==='1'){
      // 5.0.40 guarda las fotos nuevas respecto del cuerpo util y conserva el
      // calculo historico para coordenadas libres creadas por versiones previas.
      const rawPage=String(figure.dataset.sglPage||'').trim(),pageIndex=rawPage?Math.max(1,Math.round(Number(rawPage)||1)):null,rawLeft=Math.max(0,Number(figure.dataset.sglLeft||0)||0),rawTop=Math.max(0,Number(figure.dataset.sglTop||0)||0),left=pageIndex?rawLeft*.75:(rawLeft*.75)-85.04,top=pageIndex?rawTop*.75+(pageIndex-1)*(792-24-42.52-62):Math.max(0,(rawTop*.75)-12),z=Math.max(1,Math.min(9999,Math.round(Number(figure.dataset.sglZ||1)||1)));
      figure.setAttribute('style',`position:absolute;float:none;clear:none;left:${left.toFixed(2)}pt;top:${top.toFixed(2)}pt;width:${scale}%;max-width:100%;box-sizing:border-box;margin:0;text-align:left;z-index:${z};page-break-inside:avoid`);
      const img=figure.querySelector('img');if(img)img.setAttribute('style','display:block;width:100%;max-width:100%;max-height:260pt;object-fit:contain');
      continue;
    }
    const fallbackX=figure.dataset.sglAlign==='left'?0:figure.dataset.sglAlign==='right'?100:50,rawX=Number(figure.dataset.sglX),x=Math.max(0,Math.min(100,Number.isFinite(rawX)?rawX:fallbackX)),left=Math.max(0,100-scale)*(x/100),right=Math.min(100,left+scale),newRow=!inPhotoRun||left+.15<(previousRight??0),gap=Math.max(0,newRow?left:left-(previousRight??0));
    figure.setAttribute('style',`float:left;clear:${newRow?'left':'none'};width:${scale}%;max-width:100%;box-sizing:border-box;margin:6pt 0 6pt ${gap.toFixed(3)}%;text-align:left;page-break-inside:avoid`);
    const img=figure.querySelector('img');if(img)img.setAttribute('style','display:block;width:100%;max-width:100%;max-height:260pt;object-fit:contain');previousRight=right;inPhotoRun=true;
  }
}
function sgl512InlineReportHtml(html,attached=[]){const box=document.createElement('div');box.innerHTML=sanitizeRichText(html||'');const files=new Map((attached||[]).map(file=>[String(file.id||''),file]));for(const figure of [...box.querySelectorAll('figure[data-sgl-photo-id]')]){const id=String(figure.dataset.sglPhotoId||''),file=files.get(id);if(!file||file.kind!=='photo'||!file.dataUrl){figure.remove();continue;}figure.dataset.sglScale=String(Math.max(1,Math.min(100,Number(figure.dataset.sglScale||file.reportScale||70)||70)));let img=figure.querySelector('img');if(!img){img=document.createElement('img');figure.appendChild(img);}img.setAttribute('src',file.dataUrl);img.setAttribute('alt','Evidencia fotográfica');}sgl514LayoutInlineReportFigures(box);return box.innerHTML;}
function sgl512RichReportBlocks(html,fallback=''){
  const container=document.createElement('div');container.innerHTML=sanitizeRichText(html||plainTextToRichHtml(fallback));
  const blocks=[];
  const styleFlags=(node,current)=>{
    const tag=node.tagName||'',st=node.style||{},weight=String(st.fontWeight||''),decoration=String(st.textDecorationLine||st.textDecoration||''),size=parseFloat(st.fontSize),raw=String(st.fontSize||'');
    return {...current,
      bold:weight?weight==='bold'||Number(weight)>=600:current.bold||['B','STRONG'].includes(tag),
      italic:st.fontStyle?st.fontStyle==='italic'||st.fontStyle==='oblique':current.italic||['I','EM'].includes(tag),
      underline:decoration?decoration.includes('underline'):current.underline||tag==='U',
      ...(size?{size:/px$/.test(raw)?size*.75:size}:{}),
      ...(st.fontFamily?{family:st.fontFamily}:{}),
      ...(sgl532PdfColor(st.color)?{color:sgl532PdfColor(st.color)}:{}),
      ...(sgl532PdfColor(st.backgroundColor)?{highlight:sgl532PdfColor(st.backgroundColor)}:{})};
  };
  const runsFor=(root,style={})=>{const runs=[];const walk=(node,current)=>{if(node.nodeType===3){if(node.nodeValue)runs.push({text:node.nodeValue.replace(/\u200b/g,''),...current});return;}if(node.nodeType!==1)return;if(node.tagName==='FIGURE')return;const next=styleFlags(node,current);if(node.tagName==='BR')runs.push({text:'\n',hardBreak:true,...next});else [...node.childNodes].forEach(child=>walk(child,next));};walk(root,style);return runs;};
  const meaningful=runs=>runs.some(run=>String(run.text||'').replace(/\s+/g,'').length>0);
  const blockAlign=node=>['left','center','right','justify'].includes(String(node?.style?.textAlign||'').toLowerCase())?String(node.style.textAlign).toLowerCase():'left';
  let inlineRuns=[];const flushInline=()=>{if(meaningful(inlineRuns))blocks.push({type:'text',runs:inlineRuns,align:'left'});inlineRuns=[];};
  for(const node of [...container.childNodes]){
    if(node.nodeType===1&&node.tagName==='TABLE'){
      flushInline();const cols=[...node.querySelectorAll(':scope > colgroup > col')],occupied=[],rows=[];let count=cols.length;
      for(const [index,row] of [...node.rows].entries()){const cells=[];let col=0;for(const cell of row.cells){while(occupied[col]>index)col++;const span=cell.colSpan||1,rowspan=cell.rowSpan||1;cells.push({col,span,rowspan,align:blockAlign(cell),vertical:cell.style.verticalAlign,blocks:sgl512RichReportBlocks(cell.innerHTML,'')});for(let n=col;n<col+span;n++)occupied[n]=index+rowspan;col+=span;count=Math.max(count,col);}rows.push({cells,height:(parseFloat(row.style.height)||0)*.75});}
      blocks.push({type:'table',rows,weights:Array.from({length:Math.max(1,count)},(_,i)=>parseFloat(cols[i]?.style.width)||1)});continue;
    }
    if(node.nodeType===1&&node.tagName==='FIGURE'&&node.dataset.sglPhotoId){
      flushInline();const align=['left','center','right'].includes(node.dataset.sglAlign)?node.dataset.sglAlign:'center',rawX=Number(node.dataset.sglX),x=Number.isFinite(rawX)?Math.max(0,Math.min(100,rawX)):(align==='left'?0:align==='right'?100:50),free=node.dataset.sglFree==='1',left=Math.max(0,Number(node.dataset.sglLeft||0)||0),top=Math.max(0,Number(node.dataset.sglTop||0)||0),z=Math.max(1,Math.min(9999,Math.round(Number(node.dataset.sglZ||1)||1))),rawPage=String(node.dataset.sglPage||'').trim(),page=rawPage?Math.max(1,Math.min(999,Math.round(Number(rawPage)||1))):null,height=Math.max(0,Number(node.dataset.sglHeight||0)||0);blocks.push({type:'photo',id:String(node.dataset.sglPhotoId),scale:Math.max(1,Math.min(100,Number(node.dataset.sglScale||70)||70)),align,x,free,left,top,z,page,height});continue;
    }
    if(node.nodeType===1&&['UL','OL'].includes(node.tagName)){
      flushInline();let index=1;for(const item of [...node.children])if(item.tagName==='LI'){
        const prefix=node.tagName==='OL'?`${index++}. `:node.style.listStyleType==='circle'?'○ ':node.style.listStyleType==='square'?'▪ ':'• ';
        blocks.push({type:'text',runs:[{text:node.tagName==='UL'?'• ':prefix,marker:node.tagName==='UL'?node.style.listStyleType||'disc':''},...runsFor(item)],align:blockAlign(item)});
      }continue;
    }
    if(node.nodeType===1&&['P','DIV'].includes(node.tagName)){
      flushInline();const runs=runsFor(node);if(meaningful(runs))blocks.push({type:'text',runs,align:blockAlign(node)});continue;
    }
    // Word puede pegar una frase como varios <span>/<b> hermanos. No se deben
    // convertir en párrafos separados porque eso parte títulos a mitad de línea.
    inlineRuns.push(...runsFor(node));
  }
  flushInline();
  if(!blocks.length&&fallback)blocks.push({type:'text',runs:[{text:fallback}],align:'left'});return blocks;
}
function readRichEditor(){const editor=document.getElementById('req_text');if(!editor)return {text:'',html:''};const html=sanitizeRichText(editor.innerHTML),text=reportPlainTextFromHtml(html);return {text,html:reportHtmlFromSubmission({text,richText:html})};}

function openRequirementConfig(requirementId) {
  const requirement=flatRequirements.find(item=>item.id===requirementId), config=configFor(modalOwnerId,requirement);
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>Configurar ${esc(requirementId)}</h3><p>Indicaciones visibles para este Aspirante.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><form onsubmit="saveRequirementConfig(event,'${requirementId}')"><div class="modal-body"><label class="field-group">Modalidad<input id="cfg_mode" class="field" value="${esc(config.mode)}"></label><label class="field-group">Indicaciones<textarea id="cfg_instructions" class="field">${esc(config.instructions)}</textarea></label><label class="field-group">Criterio de evaluación<textarea id="cfg_evaluation" class="field">${esc(config.evaluation)}</textarea></label></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="openRequirement('${requirementId}')">Volver</button><button class="btn primary">Guardar</button></div></form></div>`);
}

async function saveRequirementConfig(event,requirementId) {
  event.preventDefault(); database.configs[`${modalOwnerId}:${requirementId}`]={...configFor(modalOwnerId,flatRequirements.find(item=>item.id===requirementId)),mode:value('cfg_mode'),instructions:value('cfg_instructions'),evaluation:value('cfg_evaluation')}; await persist(true,{action:'Requisito configurado',entity:requirementId,detail:userById(modalOwnerId)?.fullName||''}); await openRequirement(requirementId); toast('Configuración guardada.');
}

function pendingReviews(user=currentUser()) {
  if (!user) return [];
  return accessibleAspirants(user).flatMap(aspirant=>Object.entries(ensureProgress(aspirant.id).status).filter(([,status])=>status==='in_review').map(([requirementId])=>({aspirant,requirement:flatRequirements.find(item=>item.id===requirementId)})));
}

function renderReviews() {
  const user=currentUser(), reviews=pendingReviews(user); setTop('Revisiones',ROLE_LABELS[user.rawRole624||user.role]);
  setContent(`<div class="page-head"><div><div class="eyebrow">Bandeja de trabajo</div><h1 class="page-title">${reviews.length} requisito${reviews.length===1?'':'s'} en revisión</h1><p class="subtitle">La aprobación registra responsable, fecha y estado.</p></div></div><div class="list">${reviews.map(item=>`<div class="list-item clickable" onclick="activeAspirantId='${item.aspirant.id}';openRequirement('${item.requirement.id}')">${requirementAreaBadge(item.requirement)}<div><h4>${esc(item.requirement.official_text)}</h4><div class="chips"><span class="chip">${esc(item.aspirant.fullName)}</span><span class="chip blue">${esc(clubFor(item.aspirant)?.name||'Sin Club')}</span></div></div><span class="status in_review">En revisión</span></div>`).join('')||'<div class="card empty">No hay entregas pendientes.</div>'}</div>`);
}

function reviewRequirement(status) {
  const submission=submissionFor(modalOwnerId,modalRequirementId), feedback=status==='needs_changes'?prompt('Indica qué debe corregir el Aspirante:',''):'';
  if (status==='needs_changes'&&!feedback) return;
  submission.feedback=feedback||''; submission.approval=status==='approved'?{reviewerId:currentUser().id,reviewerName:currentUser().fullName,at:isoNow()}:null; submission.updatedAt=isoNow(); ensureProgress(modalOwnerId).status[modalRequirementId]=status;
  persist(true,{action:status==='approved'?'Requisito revisado':'Corrección solicitada',entity:modalRequirementId,detail:userById(modalOwnerId)?.fullName||''}).then(()=>{closeModal();go('reviews');toast(status==='approved'?'Requisito revisado.':'Corrección enviada.');});
}

function renderBible() {
  const aspirant=targetAspirant(), bibleRequirements=flatRequirements.filter(item=>item.kind==='bible_reading'), versions=database.bibleVersions||[], selected=currentUser().bibleVersionId||''; setTop('Lectura bíblica','Diario de lectura');
  if (!aspirant) return setContent('<div class="card empty">No hay Aspirante seleccionado.</div>');
  const cards=bibleRequirements.map(req=>{const readings=ensureProgress(aspirant.id).readings[req.id]||{},done=Object.values(readings).filter(item=>item.completed).length,total=req.reading_plan.length,percent=Math.round(done*100/total);return `<div class="card"><div class="card-head"><h3>${esc(req.id)}</h3><span class="status ${done===total?'approved':'in_progress'}">${done}/${total}</span></div><p class="subtitle">${esc(req.official_text)}</p><div class="bar"><i style="width:${percent}%"></i></div><p class="muted" style="font-size:11px">${percent}% de lecturas registradas</p><button class="btn dark block" onclick="openBibleRequirement('${req.id}')">Abrir plan de lectura</button></div>`;}).join('');
  setContent(`<div class="page-head"><div><div class="eyebrow">Lectura y reflexión</div><h1 class="page-title">Plan bíblico</h1><p class="subtitle">Cada lectura guarda fecha, versículo favorito y una reflexión de uno a tres párrafos.</p></div>${contextSelector()}</div><div class="card" style="margin-bottom:16px"><div class="card-head"><div><h3>Versión bíblica en este dispositivo</h3><p class="subtitle" style="margin:0">El instalador permite cargar una versión autorizada. También puedes importar TXT/VPL o JSON aquí.</p></div><div class="actions"><select id="bible_version" class="field" onchange="selectBibleVersion(this.value)" style="margin:0;min-width:220px"><option value="">Sin versión instalada</option>${versions.map(version=>`<option value="${version.id}" ${selected===version.id?'selected':''}>${esc(version.name)}</option>`).join('')}</select><label class="btn primary">Cargar versión<input type="file" hidden accept=".txt,.vpl,.json,text/plain,application/json" onchange="importBibleVersion(this.files[0]);this.value=''"/></label></div></div></div><div class="grid3">${cards}</div><div class="card" style="margin-top:16px"><div class="card-head"><div><h3>Biblia en orden canónico</h3><p class="subtitle">Antiguo y Nuevo Testamento, libro por libro.</p></div></div><div class="canon-grid">${BIBLE_CANON.map(([code,name,chapters],index)=>`<button class="canon-book" onclick="openCanonicalBook('${code}')"><b>${index+1}. ${esc(name)}</b><span>${chapters} capítulo${chapters===1?'':'s'}</span></button>`).join('')}</div></div>`);
}

async function selectBibleVersion(id) { currentUser().bibleVersionId=id||null; await persist(true,{action:'Versión bíblica seleccionada',entity:id||'none',detail:database.bibleVersions.find(item=>item.id===id)?.name||''}); toast('Versión bíblica seleccionada.'); }

function bibleBookCode(name) {
  const key=normalize(name).normalize('NFD').replace(/[\u0300-\u036f]/g,'').replace(/\./g,'').replace(/\s+/g,' ').trim();
  const aliases={gen:'GEN',genesis:'GEN',exo:'EXO',exodo:'EXO',lev:'LEV',levitico:'LEV',num:'NUM',numeros:'NUM',deut:'DEU',deuteronomio:'DEU',jos:'JOS',josue:'JOS',juec:'JDG',jueces:'JDG',rut:'RUT','1 sam':'1SA','1 samuel':'1SA','2 sam':'2SA','2 samuel':'2SA','1 rey':'1KI','1 reyes':'1KI','2 rey':'2KI','2 reyes':'2KI','1 cron':'1CH','1 cronicas':'1CH','2 cron':'2CH','2 cronicas':'2CH',esd:'EZR',esdras:'EZR',neh:'NEH',nehemias:'NEH',est:'EST',ester:'EST',job:'JOB',sal:'PSA',salmos:'PSA',prov:'PRO',proverbios:'PRO',ecl:'ECC',eclesiastes:'ECC',isa:'ISA',isaias:'ISA',jer:'JER',jeremias:'JER',dan:'DAN',daniel:'DAN',amos:'AMO',abdias:'OBA',jon:'JON',jonas:'JON',miq:'MIC',miqueas:'MIC',hag:'HAG',hageo:'HAG',zac:'ZEC',zacarias:'ZEC',mal:'MAL',malaquias:'MAL',mat:'MAT',mateo:'MAT',mar:'MRK',marcos:'MRK',luc:'LUK',lucas:'LUK',juan:'JHN',hech:'ACT',hechos:'ACT',rom:'ROM','1 cor':'1CO','2 cor':'2CO',gal:'GAL',galatas:'GAL',efe:'EPH',efesios:'EPH',fil:'PHP',file:'PHM',filemon:'PHM',col:'COL','1 tes':'1TH','2 tes':'2TH','1 tim':'1TI','2 tim':'2TI',tit:'TIT',tito:'TIT',heb:'HEB',hebreos:'HEB',sant:'JAS',santiago:'JAS','1 ped':'1PE','2 ped':'2PE','1 juan':'1JN','2 juan':'2JN','3 juan':'3JN',jud:'JUD',judas:'JUD',apoc:'REV',apocalipsis:'REV'};
  const extra={cant:'SNG',cantares:'SNG','cantar de los cantares':'SNG',lam:'LAM',lamentaciones:'LAM',eze:'EZK',ezequiel:'EZK',ose:'HOS',oseas:'HOS',joel:'JOL',nah:'NAM',nahum:'NAM',hab:'HAB',habacuc:'HAB',sof:'ZEP',sofonias:'ZEP',filipenses:'PHP','1 tesalonicenses':'1TH','2 tesalonicenses':'2TH','1 timoteo':'1TI','2 timoteo':'2TI','1 pedro':'1PE','2 pedro':'2PE'};
  return extra[key]||aliases[key]||key.toUpperCase().replace(/\s/g,'');
}

async function openCanonicalBook(code) {
  const [,,chapterCount]=BIBLE_CANON.find(item=>item[0]===code)||[];const version=database.bibleVersions.find(item=>item.id===currentUser().bibleVersionId);if(!version)return toast('Primero carga o selecciona una versión bíblica.');const data=(await allFiles()).find(item=>item.id===version.fileId),canonical=BIBLE_CANON.find(item=>item[0]===code);if(!data?.verses||!canonical)return;
  showModal(`<div class="modal"><div class="modal-head"><div><h3>${esc(canonical[1])}</h3><p>${esc(version.name)} · orden canónico</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="chapter-tabs">${Array.from({length:chapterCount},(_,index)=>`<button class="btn ghost small" onclick="showCanonicalChapter('${code}',${index+1})">${index+1}</button>`).join('')}</div><div id="canonical_text" class="bible-text" style="margin-top:12px"></div></div></div>`);showCanonicalChapter(code,1,data);
}
async function showCanonicalChapter(code,chapter,provided=null){const version=database.bibleVersions.find(item=>item.id===currentUser().bibleVersionId),data=provided||(await allFiles()).find(item=>item.id===version?.fileId),root=document.getElementById('canonical_text'),book=BIBLE_CANON.find(item=>item[0]===code);if(!root||!data)return;const rows=Object.entries(data.verses).map(([key,text])=>{const [b,c,v]=key.split('|');return {b,c:Number(c),v:Number(v),text};}).filter(x=>x.b===code&&x.c===chapter).sort((a,b)=>a.v-b.v);root.innerHTML=`<h3>${esc(book?.[1])} ${chapter}</h3>${rows.map(row=>`<span class="verse"><sup>${row.v}</sup>${esc(row.text)}</span> `).join('')||'<p>Este capítulo no aparece en la versión cargada.</p>'}`;}

async function importBibleVersion(file) {
  if(!file)return;
  try {
    const text=await file.text(); let name=file.name.replace(/\.(json|txt|vpl)$/i,''),licenseNote='Copia cargada por el usuario',verses={};
    if(file.name.toLowerCase().endsWith('.json')) {
      const parsed=JSON.parse(text); name=parsed.name||name; licenseNote=parsed.license||licenseNote;
      const source=parsed.verses||parsed;
      for(const [reference,verseText] of Object.entries(source)) { const match=reference.match(/^((?:[123]\s*)?[^\d]+?)\s+(\d+):(\d+)$/); if(match&&typeof verseText==='string')verses[`${bibleBookCode(match[1])}|${Number(match[2])}|${Number(match[3])}`]=verseText.trim(); }
    } else {
      for(const line of text.split(/\r?\n/)) { const match=line.trim().match(/^((?:[123]\s*)?[A-Za-zÁÉÍÓÚÜÑáéíóúüñ. ]+?)\s+(\d+):(\d+)\s+(.+)$/); if(match)verses[`${bibleBookCode(match[1])}|${Number(match[2])}|${Number(match[3])}`]=match[4].trim(); }
    }
    const count=Object.keys(verses).length; if(count<10)throw new Error('No se reconocieron suficientes versículos. Usa el formato “Génesis 1:1 Texto”.');
    const chosen=prompt('Nombre que aparecerá para esta versión:',name); if(!chosen)return;
    const id=newId('bible'),fileId=newId('file'); await filePut({id:fileId,ownerId:currentUser().id,kind:'bible_version',name:chosen,type:'application/json',size:text.length,verses,createdAt:isoNow()});
    database.bibleVersions.push({id,name:chosen,fileId,license:licenseNote,verseCount:count,installedAt:isoNow(),installedBy:currentUser().id}); currentUser().bibleVersionId=id;
    await persist(true,{action:'Versión bíblica cargada',entity:id,detail:`${chosen} · ${count} versículos`}); renderBible(); toast(`${chosen}: ${count} versículos disponibles sin conexión.`);
  } catch(error){toast(`No se pudo cargar la versión: ${error.message}`);}
}

function readingReferenceParts(passage) {
  const match=String(passage).match(/^((?:[123]\s*)?[^\d]+?)\s*(.+)$/); if(!match)return {book:'',chapters:[]}; const book=bibleBookCode(match[1]),rest=match[2],chapters=[];
  const first=rest.match(/^(\d+)/); if(first)chapters.push(Number(first[1]));
  if(!rest.split(/[;y]/)[0].includes(':')) for(const extra of rest.matchAll(/[,;y]\s*(\d+)(?!\s*[-,])/g))chapters.push(Number(extra[1]));
  for(const extra of rest.matchAll(/;\s*(\d+)/g))chapters.push(Number(extra[1]));
  if(rest.includes(':')) { const finalChapter=rest.match(/\by\s*(\d+)\s*$/i); if(finalChapter)chapters.push(Number(finalChapter[1])); }
  return {book,chapters:[...new Set(chapters)]};
}

async function bibleTextForPassage(passage) {
  const version=database.bibleVersions.find(item=>item.id===currentUser().bibleVersionId); if(!version)return {name:'Sin versión instalada',html:'<p>Carga una versión TXT/VPL o JSON desde el módulo Lectura bíblica.</p>'};
  const data=(await allFiles()).find(item=>item.id===version.fileId),parts=readingReferenceParts(passage); if(!data?.verses)return {name:version.name,html:'<p>No se encontró el archivo de esta versión.</p>'};
  const rows=Object.entries(data.verses).map(([key,text])=>{const [book,chapter,verse]=key.split('|');return {book,chapter:Number(chapter),verse:Number(verse),text};}).filter(item=>item.book===parts.book&&parts.chapters.includes(item.chapter)).sort((a,b)=>a.chapter-b.chapter||a.verse-b.verse);
  return {name:version.name,html:rows.length?rows.map(item=>`<span class="verse" onclick="selectBibleVerse('${esc(`${passage.split(/\d/)[0].trim()} ${item.chapter}:${item.verse}`)}',this)"><sup>${item.chapter}:${item.verse}</sup>${esc(item.text)}</span> `).join(''):'<p>Esta copia no contiene el pasaje solicitado o usa nombres de libros no reconocidos.</p>'};
}

function selectBibleVerse(reference,root) { document.querySelectorAll('.verse.selected').forEach(item=>item.classList.remove('selected'));root?.classList.add('selected');const input=document.getElementById('reading_verse');if(input)input.value=reference; }

function openBibleRequirement(requirementId) {
  const aspirant=targetAspirant(), req=flatRequirements.find(item=>item.id===requirementId), readings=ensureProgress(aspirant.id).readings[req.id]||{}; modalOwnerId=aspirant.id; modalRequirementId=req.id;
  const entries=req.reading_plan.map(item=>{const entry=readings[item.id];return `<div class="reading-item ${entry?.completed?'reading-done':''}"><div><h4>${esc(item.passage)}</h4><p>${entry?.completed?`${formatDate(entry.date)} · ${esc(entry.favoriteVerse||'Sin versículo indicado')}`:'Pendiente'}</p></div><button class="btn ${entry?.completed?'ghost':'primary'} small" onclick="openReading('${item.id}')">${entry?.completed?'Editar':'Registrar'}</button></div>`;}).join('');
  const done=Object.values(readings).filter(item=>item.completed).length;
  showModal(`<div class="modal"><div class="modal-head"><div><h3>${esc(req.areaTitle)}</h3><p><b>${esc(req.official_text)}</b><br><span class="muted">Plan de lectura · ${done} de ${req.reading_plan.length} lecturas completadas</span></p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="reading-list">${entries}</div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button>${done===req.reading_plan.length?`<button class="btn primary" onclick="completeBibleRequirement('${req.id}')">Generar informe acumulado</button>`:''}</div></div>`);
}

async function openReading(readingId) {
  const aspirant=userById(modalOwnerId), req=flatRequirements.find(item=>item.id===modalRequirementId), plan=req.reading_plan.find(item=>item.id===readingId), readings=ensureProgress(aspirant.id).readings[req.id]||{}, entry=readings[readingId]||{};
  const bible=await bibleTextForPassage(plan.passage);
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>${esc(plan.passage)}</h3><p>Diario de lectura · ${esc(bible.name)}</p></div><button class="modal-close" onclick="openBibleRequirement('${req.id}')">×</button></div><form onsubmit="saveReading(event,'${readingId}')"><div class="modal-body"><div class="bible-text"><b>${esc(plan.passage)}</b><div id="bible_copy">${bible.html}</div></div><div class="actions" style="margin:10px 0"><a class="btn ghost small" target="_blank" rel="noopener" href="https://ebible.org/spaRV1909/">Abrir RV1909 en línea</a></div><label class="field-group">Fecha de lectura<input id="reading_date" class="field" type="date" required value="${esc(entry.date||localDate())}"></label><label class="field-group">Versículo que más te gustó<input id="reading_verse" class="field" required placeholder="Haz clic en un versículo o escríbelo" value="${esc(entry.favoriteVerse||'')}"></label><label class="field-group">¿Qué fue lo que más te llamó la atención?<textarea id="reading_reflection" class="field" required minlength="40" placeholder="Escribe de uno a tres párrafos…">${esc(entry.reflection||'')}</textarea></label></div><div class="modal-foot"><button type="button" class="btn ghost" onclick="openBibleRequirement('${req.id}')">Volver</button><button class="btn primary">Guardar lectura</button></div></form></div>`);
}

async function saveReading(event,readingId) {
  event.preventDefault(); const progress=ensureProgress(modalOwnerId); progress.readings[modalRequirementId]||={}; progress.readings[modalRequirementId][readingId]={completed:true,date:value('reading_date'),favoriteVerse:value('reading_verse').trim(),reflection:value('reading_reflection').trim(),updatedAt:isoNow()}; progress.status[modalRequirementId]='in_progress'; progress.updatedAt=isoNow(); await persist(true,{action:'Lectura bíblica registrada',entity:readingId,detail:modalRequirementId}); openBibleRequirement(modalRequirementId); toast('Lectura guardada y agregada al informe.');
}

async function completeBibleRequirement(requirementId) {
  const req=flatRequirements.find(item=>item.id===requirementId), readings=ensureProgress(modalOwnerId).readings[requirementId]||{}, submission=submissionFor(modalOwnerId,requirementId);
  submission.text=req.reading_plan.map(item=>{const entry=readings[item.id];return `${item.passage} — ${formatDate(entry.date)}\nVersículo: ${entry.favoriteVerse}\n${entry.reflection}`;}).join('\n\n'); submission.richText=plainTextToRichHtml(submission.text); submission.completionDate=Object.values(readings).map(item=>item.date).sort().at(-1)||localDate(); submission.updatedAt=isoNow(); ensureProgress(modalOwnerId).status[requirementId]='in_review'; await persist(true,{action:'Informe bíblico generado',entity:requirementId,detail:userById(modalOwnerId)?.fullName||''}); closeModal(); render(); toast('Informe acumulado generado y enviado a revisión.');
}

function renderReports() {
  const aspirant=targetAspirant(); setTop('Informes y documentos','Expediente');
  if (!aspirant) return setContent('<div class="card empty">No hay Aspirante seleccionado.</div>');
  modalOwnerId=aspirant.id;
  const progress=ensureProgress(aspirant.id);
  const items=flatRequirements.filter(req=>{const submission=progress.submissions[req.id]; return submission&&(submission.text||submission.attachmentIds?.length);});
  setContent(`<div class="page-head"><div><div class="eyebrow">${esc(aspirant.fullName)}</div><h1 class="page-title">Informes y revisados</h1><p class="subtitle">Los archivos se agrupan por código corto del área y conservan los logos del informe.</p></div>${contextSelector()}</div><div class="list">${items.map(req=>{const sub=submissionFor(aspirant.id,req.id),status=progress.status[req.id];return `<div class="list-item">${requirementAreaBadge(req)}<div><h4>${esc(reportBaseName(req))}</h4><p>${esc(req.official_text)}</p><div class="chips"><span class="chip">${formatDate(sub.completionDate)}</span><span class="status ${status}">${STATUS_LABELS[status]}</span>${sub.selections.map(choice=>`<span class="chip gold">${esc(choice)}</span>`).join('')}</div></div><div class="actions"><button class="btn ghost small" onclick="activeAspirantId='${aspirant.id}';openRequirement('${req.id}')">Ver</button><button class="btn primary small" onclick="exportPdf('${req.id}')">Exportar PDF</button><button class="btn ghost small" onclick="printRequirement('${req.id}')">Imprimir</button><button class="btn ghost small" onclick="exportWord('${req.id}')">Exportar a Word</button>${['aspirant','pastor_aspirant'].includes(currentUser().role)?`<button class="btn green small" onclick="sendReportToDrive('${req.id}')">Enviar a Drive</button>`:''}</div></div>`;}).join('')||'<div class="card empty">Todavía no hay informes ni certificados guardados.</div>'}</div>`);
}

async function printableRequirement(requirementId) {
  const aspirant=userById(modalOwnerId)||targetAspirant(), requirement=flatRequirements.find(item=>item.id===requirementId), baseSubmission=submissionFor(aspirant.id,requirementId), submission=liveRequirementSubmission(baseSubmission,aspirant.id,requirementId), club=clubFor(aspirant), files=await allFiles(), attached=files.filter(file=>(submission.attachmentIds||[]).includes(file.id)&&(file.kind==='photo'||String(file.type||'').startsWith('image/')));
  const officialLogo=await loadMandatoryLogo();
  const selectionHtml=submission.selections.map(item=>`<p class="chosen">${esc(item)}</p>`).join('');
  const rawTextHtml=reportHtmlFromSubmission(submission),inlinePhotoIds=sgl512InlinePhotoIds(rawTextHtml),textHtml=sgl512InlineReportHtml(rawTextHtml,attached);
  const attachments=attached.filter(file=>file.kind==='photo'&&!inlinePhotoIds.has(String(file.id))).map(file=>`<figure class="place-${file.reportPlacement||'auto'}" style="--photo-scale:${file.reportScale||70}"><img src="${file.dataUrl}" alt="Evidencia fotográfica"></figure>`).join('');
  const identity=club?.logo?aspirant.fullName:`${aspirant.fullName} · ${club?.name||'Ministerio de Conquistadores'}`,mission=reportMissionName({aspirant});
  return {aspirant,requirement,submission,club,attached,officialLogo,identity,mission,reportTitle:'INFORME CLASE AGRUPADA',html:`<header><img src="${officialLogo}"><div class="report-header-center">${mission?`<div class="report-mission">${esc(mission)}</div>`:''}<h1>INFORME CLASE AGRUPADA</h1><p>${esc(identity)}</p></div>${club?.logo?`<img src="${club.logo}">`:'<span></span>'}</header><main><h2 class="area-title">${esc(requirement.areaTitle)}</h2><h3 class="requirement-title">${esc(requirement.official_text)}</h3><p class="requirement-fill-date">${esc(reportRequirementFillDate({submission}))}</p>${selectionHtml}${textHtml?`<article>${textHtml}</article>`:''}<section class="report-photos">${attachments}</section></main>`};
}

async function loadMandatoryLogo() { if(mandatoryLogoDataUrl)return mandatoryLogoDataUrl; try{const response=await fetch(DEFAULT_REPORT_LOGO);if(!response.ok)throw new Error(`Logo ${response.status}`);mandatoryLogoDataUrl=await fileToDataUrl(await response.blob());}catch{mandatoryLogoDataUrl=appLogo();}return mandatoryLogoDataUrl; }

function reportPrintDocument(data,orientation='auto',font='Arial'){
  const selected=orientation==='landscape'?'landscape':'portrait',family=font==='Times New Roman'?'Times New Roman':'Arial';
  return `<!doctype html><html lang="es"><head><meta charset="utf-8"><title></title><style>@page{size:letter ${selected};margin:0}*{box-sizing:border-box}html,body{margin:0;padding:0;background:#fff}body{font-family:${family},sans-serif;font-size:12pt;line-height:1.3;color:#111;padding:.78cm 2cm 1.5cm 3cm}header{position:relative;display:grid;grid-template-columns:43pt 1fr 43pt;gap:10pt;align-items:center;height:56pt;border:0;text-align:center;padding:0 0 6pt;margin:0 0 10pt}header:after{content:'';position:absolute;left:53pt;right:53pt;bottom:0;border-bottom:1.8pt solid #1f739e;pointer-events:none}header img{width:43pt;height:43pt;object-fit:contain}.report-header-center{height:43pt;display:flex;flex-direction:column;justify-content:center;min-width:0}.report-mission{font-size:8.5pt;line-height:10pt;font-weight:700;color:#111;text-transform:uppercase;margin:0 0 1pt}header h1{font-size:13.5pt;line-height:16pt;color:#1f739e;margin:0}header p{font-size:10.5pt;line-height:13pt;margin:1pt 0 0}.area-title{font-size:10.5pt;line-height:14pt;color:#1f739e;text-transform:uppercase;margin:0 0 5pt}.requirement-title{font-size:11.5pt;line-height:16pt;margin:0}.requirement-fill-date{font-weight:700;color:#1f739e;margin:0;text-align:right;text-transform:none}.chosen{font-weight:bold;margin:0}article{margin-top:0}article{position:relative;min-height:1pt}article::after{content:"";display:block;clear:both}article>:is(p,div):first-child:empty,article>:is(p,div):first-child:has(>br:only-child){display:none}article p,article div{clear:both;margin:0 0 7pt;text-align:left;text-align-last:auto;line-height:1.3}.report-photos{display:grid;grid-template-columns:1fr 1fr;gap:12pt}.report-photos figure{margin:0 0 12pt;text-align:center}.report-photos img{width:calc(var(--photo-scale,70)*1%);max-width:100%;max-height:230pt;object-fit:contain}footer.report-institution-footer{position:fixed;left:3cm;right:2cm;bottom:.55cm;font-size:8.5pt;color:#111;display:grid;grid-template-columns:1fr 1fr 1fr;gap:10pt;text-transform:uppercase}footer.report-institution-footer span:nth-child(2){text-align:center}footer.report-institution-footer span:nth-child(3){text-align:right}@media print{html,body{width:100%;min-height:100%}button{display:none}}</style></head><body>${data.html}${reportInstitutionFooterHtml(data)}</body></html>`;
}
async function openReportPrintDialog(requirementId){
  try{const data=await printableRequirement(requirementId),orientation=value('req_orientation')||data.submission.orientation||'auto',font=value('req_font')||data.aspirant.reportFont||'Arial',html=reportPrintDocument(data,orientation,font);showModal(`<div class="modal material-preview report-print-modal"><div class="modal-head"><div><h3>Vista de impresión</h3><p>Revisa el informe y abre el cuadro estándar de impresión.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><iframe id="report_print_frame" class="preview-frame"></iframe></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn ghost" onclick="exportPdf('${requirementId}')">Guardar PDF</button><button class="btn primary" onclick="printReportFrame()">Abrir cuadro de impresión</button></div></div>`);const frame=document.getElementById('report_print_frame');if(frame)frame.srcdoc=html;}catch(error){await logClientError('Impresión',error);toast(`No se pudo preparar la impresión: ${error.message}`);}
}
function printReportFrame(){const frame=document.getElementById('report_print_frame');if(!frame?.contentWindow)return toast('La vista de impresión todavía no está disponible.');if(window.SGLAndroid?.printPage){try{window.SGLAndroid.printPage();return;}catch{}}frame.contentWindow.focus();frame.contentWindow.print();}
async function printRequirement(requirementId) { return openReportPrintDialog(requirementId); }

async function exportPdf(requirementId) { try{toast('Generando PDF…');const data=await printableRequirement(requirementId),blob=await makeReportPdf(requirementId,data),result=await saveBlobWithDialog(reportFileName(data),blob,'Documento PDF|*.pdf');if(result?.ok)toast(`PDF guardado en ${result.path}.`);}catch(error){await logClientError('Exportación PDF',error);toast('No se pudo generar el PDF. El error quedó registrado en Configuración.');} }
async function saveCurrentReportPreviewPdf(){const current=window.__SGL_LAST_REPORT_PREVIEW__;if(!current?.blob)return toast('La vista PDF ya no está disponible. Vuelve a presionar Visualizar PDF.');const result=await saveBlobWithDialog(current.filename||`Informe_${localDate()}.pdf`,current.blob,'Documento PDF|*.pdf');if(result?.ok)toast(`PDF guardado en ${result.path}.`);}
async function previewRequirementPdf(requirementId){try{toast('Preparando vista previa…');const data=await printableRequirement(requirementId),blob=await makeReportPdf(requirementId,data);window.__SGL_LAST_REPORT_PREVIEW__={blob,filename:reportFileName(data),requirementId};if(materialPreviewUrl)URL.revokeObjectURL(materialPreviewUrl);materialPreviewUrl=URL.createObjectURL(blob);showModal(`<div class="modal material-preview"><div class="modal-head"><div><h3>Vista previa del informe</h3><p>${esc(data.requirement.areaTitle)} · ${esc(data.requirement.official_text)}</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><iframe class="preview-frame" src="${materialPreviewUrl}#zoom=page-width"></iframe></div><div class="modal-foot"><button class="btn ghost" onclick="openRequirement('${requirementId}')">Volver al informe</button><button class="btn ghost" onclick="printRequirement('${requirementId}')">Imprimir</button><button class="btn primary" onclick="saveCurrentReportPreviewPdf()">Guardar PDF</button></div></div>`);}catch(error){toast(`No se pudo mostrar el informe: ${error.message}`);}}
function showPrintRecovery(requirementId,message){showModal(`<div class="modal small"><div class="modal-head"><div><h3>Impresión no disponible</h3><p>Windows no abrió la lista de impresoras.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="notice warn">${esc(message)}</div><p class="subtitle" style="margin-top:14px">Puedes guardar el PDF sin perder calidad o reiniciar la aplicación para volver a intentar el cuadro de impresión.</p></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn primary" onclick="exportPdf('${requirementId}')">Guardar PDF</button></div></div>`);}

async function makeReportPdf(requirementId,provided=null) {
  const data=provided||await printableRequirement(requirementId),selected=value('req_orientation')||data.submission.orientation||'auto',font=value('req_font')||data.aspirant.reportFont||'Arial',changed=data.submission.orientation!==selected||data.aspirant.reportFont!==font;
  data.submission.orientation=selected;data.aspirant.reportFont=font;if(changed)await persist(true,{action:'Formato de informe actualizado',entity:requirementId,detail:selected});
  const attachments=data.attached.map(file=>[file.id,file.createdAt,file.reportPlacement,file.reportScale,file.size]),key=JSON.stringify([data.aspirant.id,requirementId,reportSubmissionCacheKey(data.submission),selected,font,data.club?.logo?.length||0,attachments]);
  if(reportPdfCache.has(key))return reportPdfCache.get(key);
  const pdf=await makeVectorReportPdf(data,selected,font);reportPdfCache.set(key,pdf);while(reportPdfCache.size>8)reportPdfCache.delete(reportPdfCache.keys().next().value);return pdf;
}

function drawContain(ctx,image,x,y,w,h){const ratio=Math.min(w/image.naturalWidth,h/image.naturalHeight),dw=image.naturalWidth*ratio,dh=image.naturalHeight*ratio;ctx.drawImage(image,x+(w-dw)/2,y+(h-dh)/2,dw,dh);}
function loadImage(source){return new Promise((resolve,reject)=>{const image=new Image();image.onload=()=>resolve(image);image.onerror=reject;image.src=source;});}

const PDF_FONTS={AR:'Helvetica',AB:'Helvetica-Bold',AI:'Helvetica-Oblique',ABI:'Helvetica-BoldOblique',TR:'Times-Roman',TB:'Times-Bold',TI:'Times-Italic',TBI:'Times-BoldItalic'};

function reportInstitutionInfo(data={}){
  const aspirant=data.aspirant||{};
  return {
    church:String(aspirant.church||'').trim(),
    district:String(aspirant.missionaryDistrict||'').trim(),
    region:String(aspirant.region||'').trim()
  };
}
function reportMissionName(data={}){
  const aspirant=data.aspirant||{};
  return String(aspirant.mission||'').trim().toUpperCase();
}
function reportRequirementFillDate(data={}){
  const date=data?.submission?.completionDate||localDate();
  return `Fecha de realización del requisito: ${formatDate(date)}`;
}
function reportInstitutionFooterHtml(data={}){
  const info=reportInstitutionInfo(data);
  return `<footer class="report-institution-footer"><span>IGLESIA: ${esc(info.church)}</span><span>DISTRITO: ${esc(info.district)}</span><span>REGIÓN: ${esc(info.region)}</span></footer>`;
}

function reportFontCode(family,{bold=false,italic=false}={}){const times=/Times New Roman/i.test(String(family));return `${times?'T':'A'}${bold&&italic?'BI':bold?'B':italic?'I':'R'}`;}
function win1252Byte(char){const code=char.codePointAt(0);if(code<128||(code>=160&&code<=255))return code;const map={'€':128,'‚':130,'ƒ':131,'„':132,'…':133,'†':134,'‡':135,'ˆ':136,'‰':137,'Š':138,'‹':139,'Œ':140,'Ž':142,'‘':145,'’':146,'“':147,'”':148,'•':149,'–':150,'—':151,'˜':152,'™':153,'š':154,'›':155,'œ':156,'ž':158,'Ÿ':159};return map[char]??63;}
function pdfHex(text){return `<${[...String(text||'')].map(char=>win1252Byte(char).toString(16).padStart(2,'0')).join('')}>`;}
function pdfText(page,text,x,top,size,font,underline=false,color='0 0 0'){if(!text)return;const baseline=page.h-top-size;page.commands.push(`${color} rg BT /F${font} ${size} Tf 1 0 0 1 ${x.toFixed(2)} ${baseline.toFixed(2)} Tm ${pdfHex(text)} Tj ET 0 0 0 rg`);if(underline){const width=measureReportText(text,size,font);page.commands.push(`${color} RG ${x.toFixed(2)} ${(baseline-1.5).toFixed(2)} m ${(x+width).toFixed(2)} ${(baseline-1.5).toFixed(2)} l 0.7 w S 0 0 0 RG`);}}
function measureReportText(text,size,fontCode){const canvas=measureReportText.canvas||(measureReportText.canvas=document.createElement('canvas')),ctx=canvas.getContext('2d'),times=fontCode.startsWith('T'),bold=fontCode.includes('B'),italic=fontCode.includes('I');ctx.font=`${italic?'italic ':''}${bold?'700 ':'400 '}${size}px ${times?'Times New Roman':'Arial'}`;return ctx.measureText(text).width;}
function pdfImage(page,resource,x,top,w,h){const name=`Im${page.images.length+1}`;page.images.push({...resource,name});const rad=-(Number(resource.angle)||0)*Math.PI/180,c=Math.cos(rad),sn=Math.sin(rad),cx=x+w/2,cy=page.h-top-h/2,a=w*c,b=w*sn,d=h*c,e=-h*sn,tx=cx-a/2-e/2,ty=cy-b/2-d/2;page.commands.push(`q ${a.toFixed(2)} ${b.toFixed(2)} ${e.toFixed(2)} ${d.toFixed(2)} ${tx.toFixed(2)} ${ty.toFixed(2)} cm /${name} Do Q`);}

async function jpegResource(source,maxDimension=3200,keepTransparency=false){const image=await loadImage(source),scale=Math.min(1,maxDimension/Math.max(image.naturalWidth,image.naturalHeight)),width=Math.max(1,Math.round(image.naturalWidth*scale)),height=Math.max(1,Math.round(image.naturalHeight*scale)),canvas=document.createElement('canvas'),ctx=canvas.getContext('2d',{willReadFrequently:keepTransparency});canvas.width=width;canvas.height=height;ctx.clearRect(0,0,width,height);ctx.drawImage(image,0,0,width,height);if(keepTransparency){const rgba=ctx.getImageData(0,0,width,height).data,rgb=new Uint8Array(width*height*3),alpha=new Uint8Array(width*height);let hasTransparency=false;for(let index=0,pixel=0,rgbIndex=0;index<rgba.length;index+=4,pixel+=1,rgbIndex+=3){rgb[rgbIndex]=rgba[index];rgb[rgbIndex+1]=rgba[index+1];rgb[rgbIndex+2]=rgba[index+2];alpha[pixel]=rgba[index+3];if(alpha[pixel]<254)hasTransparency=true;}return {bytes:rgb,alphaBytes:hasTransparency?alpha:null,w:width,h:height,filter:null};}ctx.fillStyle='#fff';ctx.globalCompositeOperation='destination-over';ctx.fillRect(0,0,width,height);ctx.globalCompositeOperation='source-over';const blob=await new Promise(resolve=>canvas.toBlob(resolve,'image/jpeg',.98));return {bytes:new Uint8Array(await blob.arrayBuffer()),alphaBytes:null,w:width,h:height,filter:'DCTDecode'};}
function richParagraphs(html,fallback=''){const container=document.createElement('div');container.innerHTML=sanitizeRichText(html||plainTextToRichHtml(fallback));const runsFor=(root,style={})=>{const runs=[];const walk=(node,current)=>{if(node.nodeType===3){if(node.nodeValue)runs.push({text:node.nodeValue.replace(/\u200b/g,''),...current});return;}if(node.nodeType!==1)return;const next={...current,bold:current.bold||['B','STRONG'].includes(node.tagName),underline:current.underline||node.tagName==='U',italic:current.italic||['I','EM'].includes(node.tagName)};if(node.tagName==='BR')runs.push({text:'\n',hardBreak:true,...next});else [...node.childNodes].forEach(child=>walk(child,next));};walk(root,style);return runs;};const blocks=[...container.childNodes].filter(node=>node.nodeType===1&&['P','DIV'].includes(node.tagName)),blockFor=node=>({runs:runsFor(node),align:['left','center','right','justify'].includes(node.style?.textAlign)?node.style.textAlign:'left'});if(blocks.length)return blocks.map(blockFor);return [{runs:runsFor(container),align:'left'}];}
function splitReportToken(token,size,code,maxWidth){
  const parts=[];let text='',width=0;
  for(const char of [...String(token||'')]){const charWidth=measureReportText(char,size,code);if(text&&width+charWidth>maxWidth){parts.push({text,width});text=char;width=charWidth;}else{text+=char;width+=charWidth;}}
  if(text)parts.push({text,width});return parts;
}
function wrapRichRuns(runs,maxWidth,size,family){
  // 5.0.39: reproducir el colapso de espacios de HTML/Word. El navegador
  // visualiza una secuencia de espacios, tabulaciones o saltos de texto como
  // un solo espacio; el PDF debe medir exactamente con la misma regla. Solo
  // un <br> marcado como hardBreak crea un salto de linea obligatorio.
  const lines=[];let line=[],width=0;const push=()=>{lines.push(line);line=[];width=0;};
  const append=(run,text,code,tokenWidth,runSize)=>{
    const whitespace=/^\s+$/.test(text);
    if(whitespace){if(!line.length||/^\s+$/.test(line[line.length-1]?.text||''))return;text=' ';tokenWidth=measureReportText(' ',runSize,code);}
    if(!whitespace&&line.length&&width+tokenWidth>maxWidth)push();
    if(!line.length&&whitespace)return;
    line.push({...run,text,font:code,width:tokenWidth,size:runSize});width+=tokenWidth;
  };
  for(const run of runs){
    if(run?.hardBreak){push();continue;}
    const runSize=Math.max(10,Math.min(14,Number(run?.size||size)||size)),normalized=String(run?.text||'').replace(/\u00a0/g,' ').replace(/[\t\r\n\f\v ]+/g,' '),code=reportFontCode(run.family||family,run);
    for(const token of normalized.split(/( +)/).filter(Boolean)){
      const tokenWidth=measureReportText(token,runSize,code);
      if(!/^\s+$/.test(token)&&tokenWidth>maxWidth){for(const part of splitReportToken(token,runSize,code,maxWidth))append(run,part.text,code,part.width,runSize);continue;}
      append(run,token,code,tokenWidth,runSize);
    }
  }
  if(line.length||!lines.length)push();return lines;
}

async function makeVectorReportPdf(data,orientation,fontFamily){
  // 5.0.43: el PDF replica la misma regla de la hoja visible. Una foto apaisada
  // no cambia la pagina; solo la opcion Horizontal produce Carta horizontal.
  const photos=data.attached.filter(file=>file.kind==='photo'),landscape=orientation==='landscape',pageSize=landscape?[792,612]:[612,792],m={top:24,bottom:42.52,left:85.04,right:56.7},usable=pageSize[0]-m.left-m.right,pages=[],sky='0.122 0.451 0.620',bodyFontSize=Math.max(10,Math.min(14,Number(data.submission?.reportFontSize||12)||12)),bodyLineHeight=Math.max(13,bodyFontSize*1.3);
  const official=await jpegResource(data.officialLogo,256,true),clubLogo=data.club?.logo?await jpegResource(data.club.logo,256,true):null,photoResources=new Map();let page,y,lockPageBreaks=false;
  const addPage=()=>{page={w:pageSize[0],h:pageSize[1],commands:[],images:[]};pages.push(page);const logo=46,headerTop=m.top;if(data.programId==='music_ministry'){const scale=Math.min(logo/official.w,logo/official.h),w=official.w*scale,h=official.h*scale;pdfImage(page,official,m.left+(logo-w)/2,headerTop+(logo-h)/2,w,h);}else pdfImage(page,official,m.left,headerTop,logo,logo);if(clubLogo)pdfImage(page,clubLogo,page.w-m.right-logo,headerTop,logo,logo);const title=data.reportTitle||'INFORME CLASE AGRUPADA',mission=data.mission||reportMissionName(data),titleFont=reportFontCode(fontFamily,{bold:true}),missionFont=reportFontCode(fontFamily,{bold:true}),identityFont=reportFontCode(fontFamily),titleW=measureReportText(title,13.5,titleFont),identitySize=measureReportText(data.identity,10.5,identityFont)>usable-90?9:10.5,identityW=measureReportText(data.identity,identitySize,identityFont),missionSize=8.5,missionW=measureReportText(mission,missionSize,missionFont),titleTop=mission?headerTop+11:headerTop+5,identityTop=mission?headerTop+28:headerTop+23;if(mission)pdfText(page,mission,Math.max(m.left,(page.w-missionW)/2),headerTop+1,missionSize,missionFont,false,'0 0 0');pdfText(page,title,(page.w-titleW)/2,titleTop,13.5,titleFont,false,sky);pdfText(page,data.identity,Math.max(m.left,(page.w-identityW)/2),identityTop,identitySize,identityFont);page.commands.push(`${sky} RG ${m.left.toFixed(2)} ${(page.h-headerTop-52).toFixed(2)} m ${(page.w-m.right).toFixed(2)} ${(page.h-headerTop-52).toFixed(2)} l 1.8 w S 0 0 0 RG`);const footerInfo=reportInstitutionInfo(data),footerFont=reportFontCode(fontFamily),footerTop=page.h-30,footerColor='0 0 0',footerLeft=`IGLESIA: ${footerInfo.church||''}`,footerCenter=`DISTRITO: ${footerInfo.district||''}`,footerRight=`REGIÓN: ${footerInfo.region||''}`,centerW=measureReportText(footerCenter,8.5,footerFont),rightW=measureReportText(footerRight,8.5,footerFont);pdfText(page,footerLeft,m.left,footerTop,8.5,footerFont,false,footerColor);pdfText(page,footerCenter,(page.w-centerW)/2,footerTop,8.5,footerFont,false,footerColor);pdfText(page,footerRight,page.w-m.right-rightW,footerTop,8.5,footerFont,false,footerColor);y=headerTop+62;};
  const ensure=height=>{if(!lockPageBreaks&&y+height>page.h-m.bottom)addPage();};
  const drawLines=(lines,size,lineHeight,gap=0,align='left')=>{const heights=lines.map(line=>Math.max(lineHeight,...line.map(run=>Math.max(13,(Number(run?.size||size)||size)*1.3)))),total=heights.reduce((sum,height)=>sum+height,0)+gap,remaining=Math.floor((page.h-m.bottom-y)/Math.max(1,lineHeight));if(!lockPageBreaks&&((lines.length<=6&&total>page.h-m.bottom-y)||(lines.length>1&&remaining<2)))addPage();for(let lineIndex=0;lineIndex<lines.length;lineIndex+=1){const line=lines[lineIndex],currentLineHeight=heights[lineIndex]||lineHeight;ensure(currentLineHeight);const lineWidth=line.reduce((sum,run)=>sum+(run.width||0),0),last=lineIndex===lines.length-1,spaces=line.filter(run=>/^\s+$/.test(run.text||'')).length,justifyExtra=align==='justify'&&!last&&spaces?Math.max(0,(usable-lineWidth)/spaces):0;let x=align==='center'?m.left+Math.max(0,(usable-lineWidth)/2):align==='right'?m.left+Math.max(0,usable-lineWidth):m.left;for(const run of line){const runSize=Math.max(10,Math.min(14,Number(run?.size||size)||size));if(run.highlight)page.commands.push(`q ${run.highlight} rg ${x.toFixed(2)} ${(page.h-y-runSize*1.18).toFixed(2)} ${(run.width||0).toFixed(2)} ${(runSize*1.3).toFixed(2)} re f Q`);if(run.marker&&run.text.trim()&&['circle','square'].includes(run.marker)){const cx=x+runSize*.2,cy=page.h-y-runSize*.7,r=runSize*.14,k=r*.5522848;page.commands.push(run.marker==='square'?`q 0 0 0 rg ${(cx-r).toFixed(2)} ${(cy-r).toFixed(2)} ${(2*r).toFixed(2)} ${(2*r).toFixed(2)} re f Q`:`q 0 0 0 RG .7 w ${cx+r} ${cy} m ${cx+r} ${cy+k} ${cx+k} ${cy+r} ${cx} ${cy+r} c ${cx-k} ${cy+r} ${cx-r} ${cy+k} ${cx-r} ${cy} c ${cx-r} ${cy-k} ${cx-k} ${cy-r} ${cx} ${cy-r} c ${cx+k} ${cy-r} ${cx+r} ${cy-k} ${cx+r} ${cy} c S Q`);}else pdfText(page,run.text,x,y,runSize,run.font,run.underline,run.color||'0 0 0');x+=run.width;if(justifyExtra&&/^\s+$/.test(run.text||''))x+=justifyExtra;}y+=currentLineHeight;}y+=gap;};
  const paragraph=(runs,size=12,lineHeight=18,gap=12,style={},align='left')=>{const styled=runs.map(run=>({...run,...style,bold:style.bold||run.bold,italic:style.italic||run.italic,underline:style.underline||run.underline})),lines=wrapRichRuns(styled,usable,size,fontFamily);drawLines(lines,size,lineHeight,gap,align);};
  const drawPhotoRows=async list=>{for(let index=0;index<list.length;index+=2){const row=list.slice(index,index+2),resources=[];for(const item of row){let resource=photoResources.get(item.id);if(!resource){resource=await jpegResource(item.dataUrl);resource.angle=Number(data.submission?.reportPhotoBoxes?.[item.id]?.angle)||0;photoResources.set(item.id,resource);}resources.push([item,resource]);}const gap=12,col=(usable-gap)/2,boxes=resources.map(([item,res])=>{const scale=Math.max(.01,Math.min(1,(item.reportScale||70)/100)),maxW=col*scale,maxH=Math.min(230,(page.h-m.top-m.bottom)*.46)*scale,ratio=Math.min(maxW/res.w,maxH/res.h);return {item,res,w:res.w*ratio,h:res.h*ratio};}),rowH=Math.max(...boxes.map(box=>box.h))+14;ensure(rowH);boxes.forEach((box,colIndex)=>{const colX=m.left+colIndex*(col+gap),x=colX+(col-box.w)/2;pdfImage(page,box.res,x,y,box.w,box.h);});y+=rowH;}};
  addPage();
  paragraph([{text:String(data.requirement.areaTitle||'').toUpperCase()}],10.5,14,5,{bold:true,color:sky});
  paragraph([{text:data.requirement.official_text}],11.5,16,0,{bold:true});
  {const dateText=reportRequirementFillDate(data),dateFont=reportFontCode(fontFamily,{bold:true}),dateWidth=measureReportText(dateText,10.5,dateFont);ensure(14);pdfText(page,dateText,m.left+Math.max(0,usable-dateWidth),y,10.5,dateFont,false,sky);y+=14;}
  for(const selection of data.submission.selections||[])paragraph([{text:selection}],12,18,0,{bold:true});
  // 5.0.40: origen exacto del cuerpo editable de la primera hoja. Las fotos
  // libres nuevas guardan pagina + coordenadas relativas a ese cuerpo.
  const reportBodyStartY=y;
  const layoutPages=Array.isArray(data.reportLayoutPages)&&data.reportLayoutPages.length?data.reportLayoutPages:Array.isArray(data.submission?.reportLayoutPages)?data.submission.reportLayoutPages:[];
  const layoutPhotoBoxes=data.reportPhotoBoxes||data.submission?.reportPhotoBoxes||{};
  const inlineSource=layoutPages.length?layoutPages.join(''):reportHtmlFromSubmission(data.submission);
  const inlineBlocks=sgl512RichReportBlocks(inlineSource,data.submission.text),inlineIds=new Set(inlineBlocks.filter(block=>block.type==='photo').map(block=>block.id));
  const page1=photos.filter(file=>file.reportPlacement==='page1'&&!inlineIds.has(String(file.id))),page2=photos.filter(file=>file.reportPlacement==='page2'&&!inlineIds.has(String(file.id))),automatic=photos.filter(file=>!['page1','page2'].includes(file.reportPlacement)&&!inlineIds.has(String(file.id)));
  const inlinePhotoMetrics=block=>{const item=photos.find(file=>String(file.id)===String(block.id));if(!item)return null;const scale=Math.max(.01,Math.min(1,(block.scale||item.reportScale||70)/100)),legacyX=block.align==='left'?0:block.align==='right'?100:50,position=Math.max(0,Math.min(100,Number.isFinite(Number(block.x))?Number(block.x):legacyX)),left=Math.max(0,1-scale)*(position/100);return {block,item,scale,left,right:Math.min(1,left+scale)};};
  const drawInlinePhotoRow=async row=>{const boxes=[];for(const entry of row){let resource=photoResources.get(entry.item.id);if(!resource){resource=await jpegResource(entry.item.dataUrl);photoResources.set(entry.item.id,resource);}const captured=layoutPhotoBoxes[String(entry.item.id)]||null,slotW=usable*entry.scale,maxH=Math.min(270,(page.h-m.top-m.bottom)*.5)*entry.scale,ratio=Math.min(slotW/resource.w,maxH/resource.h),capturedW=Number(captured?.width||0)*.75,w=capturedW>0?Math.min(usable,capturedW):resource.w*ratio,h=w*(resource.h/Math.max(1,resource.w));boxes.push({...entry,resource,slotW,w,h});}if(!boxes.length)return;const rowH=Math.max(...boxes.map(box=>box.h))+14;ensure(rowH);for(const box of boxes){const slotX=m.left+usable*box.left,x=slotX+Math.max(0,(box.slotW-box.w)/2);pdfImage(page,box.resource,x,y,box.w,box.h);}y+=rowH;};

  async function renderInlineBlocks5050(blocks,freeCollector,pageNumber){
    for(let index=0;index<blocks.length;){const block=blocks[index];
      if(block.type==='table'){drawTable606(block);index++;continue;}
      if(block.type==='photo'&&block.free){freeCollector.push({...block,page:pageNumber||block.page||1});index+=1;continue;}
      if(block.type!=='photo'){paragraph(block.runs,bodyFontSize,bodyLineHeight,Math.max(5,bodyFontSize*.58),{},block.align);index+=1;continue;}
      const row=[];let previousRight=-1;while(index<blocks.length&&blocks[index].type==='photo'&&!blocks[index].free){const entry=inlinePhotoMetrics(blocks[index]);if(!entry){index+=1;continue;}if(row.length&&entry.left+.001<previousRight)break;row.push(entry);previousRight=entry.right;index+=1;}await drawInlinePhotoRow(row);
    }
  }

  // Keep resized columns, minimum row heights and merged cells in vector PDFs.
  function drawTable606(table){const total=table.weights.reduce((a,b)=>a+b,0)||1,widths=table.weights.map(w=>usable*w/total),xs=[m.left];for(const w of widths)xs.push(xs.at(-1)+w);const heights=table.rows.map(row=>Math.max(12,row.height||0)),cells=[],padding=3.75;
    table.rows.forEach((row,ri)=>{for(const cell of row.cells){const width=widths.slice(cell.col,cell.col+cell.span).reduce((a,b)=>a+b,0),lines=[];for(const block of cell.blocks){if(block.type!=='text')continue;const wrapped=wrapRichRuns(block.runs,Math.max(1,width-padding*2-.75),bodyFontSize,fontFamily);for(const line of wrapped)lines.push({runs:line,align:block.align==='left'?cell.align:block.align,height:Math.max(bodyLineHeight,...line.map(r=>(r.size||bodyFontSize)*1.3))});if(lines.length)lines.at(-1).height+=3;}const height=lines.reduce((n,l)=>n+l.height,0)+padding*2+.75;cells.push({...cell,ri,width,lines,height});if(cell.rowspan===1)heights[ri]=Math.max(heights[ri],height);}});
    for(const cell of cells){const end=Math.min(heights.length,cell.ri+cell.rowspan),room=heights.slice(cell.ri,end).reduce((a,b)=>a+b,0);if(room<cell.height)heights[end-1]+=cell.height-room;}
    const tops=[];for(let ri=0;ri<heights.length;ri++){const spanning=cells.filter(c=>c.ri===ri),end=Math.max(ri+1,...spanning.map(c=>Math.min(heights.length,ri+c.rowspan)));ensure(heights.slice(ri,end).reduce((a,b)=>a+b,0));tops[ri]={page,top:y};y+=heights[ri];}
    for(const cell of cells){const target=tops[cell.ri],height=heights.slice(cell.ri,cell.ri+cell.rowspan).reduce((a,b)=>a+b,0),x=xs[cell.col],top=target.top,p=target.page;p.commands.push(`q .333 .333 .333 RG .75 w ${x.toFixed(2)} ${(p.h-top-height).toFixed(2)} ${cell.width.toFixed(2)} ${height.toFixed(2)} re S Q`);let cy=top+padding+(cell.vertical==='middle'?Math.max(0,(height-cell.height)/2):cell.vertical==='bottom'?Math.max(0,height-cell.height):0);for(const line of cell.lines){const lineW=line.runs.reduce((n,r)=>n+(r.width||0),0);let cx=x+padding+(line.align==='center'?Math.max(0,(cell.width-2*padding-lineW)/2):line.align==='right'?Math.max(0,cell.width-2*padding-lineW):0);for(const run of line.runs){pdfText(p,run.text,cx,cy,run.size||bodyFontSize,run.font,run.underline,run.color||'0 0 0');cx+=run.width||0;}cy+=line.height;}}
  }

  if(layoutPages.length){
    // 5.0.52: la pagina decidida por el editor Carta es autoritativa. El motor
    // vectorial ya no vuelve a decidir en que hoja cae cada parrafo.
    const freePhotoBlocks=[];
    lockPageBreaks=true;
    for(let pageIndex=0;pageIndex<layoutPages.length;pageIndex+=1){
      if(pageIndex>0)addPage();
      y=pageIndex===0?reportBodyStartY:m.top+62;
      const blocks=sgl512RichReportBlocks(layoutPages[pageIndex]||'','');
      await renderInlineBlocks5050(blocks,freePhotoBlocks,pageIndex+1);
    }
    lockPageBreaks=false;
    // Archivos fotograficos historicos que no estan dentro del texto se anexan
    // despues de las paginas WYSIWYG para no alterar la paginacion visible.
    const detached=[...page1,...page2,...automatic];if(detached.length){addPage();await drawPhotoRows(detached);}
    for(const block of freePhotoBlocks.sort((a,b)=>(a.z||1)-(b.z||1))){
      const item=photos.find(file=>String(file.id)===String(block.id));if(!item)continue;
      let resource=photoResources.get(item.id);if(!resource){resource=await jpegResource(item.dataUrl);resource.angle=Number(data.submission?.reportPhotoBoxes?.[item.id]?.angle)||0;photoResources.set(item.id,resource);}
      const captured=layoutPhotoBoxes[String(block.id)]||{},targetPage=Math.max(1,Number(captured.page||block.page||1)||1),targetIndex=targetPage-1;while(pages.length<=targetIndex)addPage();
      const target=pages[targetIndex],originY=targetPage===1?reportBodyStartY:m.top+62,capturedW=Math.max(0,Number(captured.width||0))*.75;
      const fallbackW=Math.max(1,usable*Math.max(.01,Math.min(1,(block.scale||item.reportScale||70)/100))),fallbackRatio=Math.min(fallbackW/resource.w,Math.min(270,(pageSize[1]-m.top-m.bottom)*.55)/resource.h),drawW=capturedW>0?capturedW:resource.w*fallbackRatio,drawH=drawW*(resource.h/Math.max(1,resource.w));
      const rawLeft=Number(captured.left??block.left??0)||0,rawTop=Number(captured.top??block.top??0)||0,minX=m.left,maxX=Math.max(minX,target.w-m.right-drawW),x=Math.max(minX,Math.min(maxX,m.left+rawLeft*.75)),maxY=Math.max(originY,target.h-m.bottom-drawH),topOnPage=originY+Math.max(0,rawTop*.75);
      pdfImage(target,resource,x,Math.max(originY,Math.min(maxY,topOnPage)),drawW,drawH);if(captured.layout==='behind')target.commands.unshift(target.commands.pop());
    }
  }else{
    // Compatibilidad con informes anteriores sin mapa de paginas 5.0.52.
    await drawPhotoRows(page1);
    if(page2.length){addPage();await drawPhotoRows(page2);}
    const freePhotoBlocks=inlineBlocks.filter(block=>block.type==='photo'&&block.free).sort((a,b)=>(a.z||1)-(b.z||1)),freeOriginPage=Math.max(0,pages.length-1),freeOriginY=y;
    await renderInlineBlocks5050(inlineBlocks,freePhotoBlocks,null);
    await drawPhotoRows(automatic);
    for(const block of freePhotoBlocks){
      const item=photos.find(file=>String(file.id)===String(block.id));if(!item)continue;
      let resource=photoResources.get(item.id);if(!resource){resource=await jpegResource(item.dataUrl);resource.angle=Number(data.submission?.reportPhotoBoxes?.[item.id]?.angle)||0;photoResources.set(item.id,resource);}
      const width=Math.max(1,usable*Math.max(.01,Math.min(1,(block.scale||item.reportScale||70)/100))),maxH=Math.min(270,(pageSize[1]-m.top-m.bottom)*.55),ratio=Math.min(width/resource.w,maxH/resource.h),drawW=resource.w*ratio,drawH=resource.h*ratio;
      if(block.page){const targetIndex=Math.max(0,block.page-1);while(pages.length<=targetIndex)addPage();const target=pages[targetIndex],originY=block.page===1?reportBodyStartY:m.top+62,topOnPage=originY+Math.max(0,block.top*.75),minX=m.left,maxX=Math.max(minX,target.w-m.right-drawW),x=Math.max(minX,Math.min(maxX,m.left+block.left*.75)),maxY=Math.max(originY,target.h-m.bottom-drawH);pdfImage(target,resource,x,Math.max(originY,Math.min(maxY,topOnPage)),drawW,drawH);}
      else{const totalTop=freeOriginY+Math.max(0,block.top*.75-12),pageOffset=Math.max(0,Math.floor(totalTop/pageSize[1])),targetIndex=freeOriginPage+pageOffset;while(pages.length<=targetIndex)addPage();const target=pages[targetIndex],topOnPage=Math.max(m.top+62,totalTop-pageOffset*pageSize[1]),x=Math.max(0,Math.min(target.w-drawW,block.left*.75));pdfImage(target,resource,x,Math.min(target.h-m.bottom-drawH,topOnPage),drawW,drawH);}
    }
  }
  const mark=data.submission?.reportWatermark,markFile=data.attached.find(f=>f.id===mark?.fileId);
  if(markFile?.dataUrl){const resource=await jpegResource(markFile.dataUrl,900,true),opacity=Math.max(.05,Math.min(.5,Number(mark.opacity)||.15));resource.alphaBytes=resource.alphaBytes||new Uint8Array(resource.w*resource.h).fill(255);resource.alphaBytes=resource.alphaBytes.map(a=>Math.round(a*opacity));
   for(const [index,target] of pages.entries())if(mark.allPages||Number(mark.page)===index+1){const ratio=Math.min(target.w*(Math.max(15,Math.min(90,Number(mark.size)||60))/100)/resource.w,target.h*.9/resource.h),w=resource.w*ratio,h=resource.h*ratio;pdfImage(target,resource,(target.w-w)/2,(target.h-h)/2,w,h);target.commands.unshift(target.commands.pop());}
  }
  return buildVectorPdf(pages);
}

function buildVectorPdf(pages){const fontEntries=Object.entries(PDF_FONTS),planPages=[];let nextId=3+fontEntries.length;for(const page of pages){const pageId=nextId++,contentId=nextId++,imageIds=[],alphaIds=[];for(const image of page.images){imageIds.push(nextId++);alphaIds.push(image.alphaBytes?nextId++:null);}planPages.push({page,pageId,contentId,imageIds,alphaIds});}const chunks=[],offsets=[0];let length=0;const add=value=>{const bytes=typeof value==='string'?new TextEncoder().encode(value):value;chunks.push(bytes);length+=bytes.length;};const object=(id,body)=>{offsets[id]=length;add(`${id} 0 obj\n`);add(body);add('\nendobj\n');};add('%PDF-1.4\n');object(1,'<< /Type /Catalog /Pages 2 0 R >>');object(2,`<< /Type /Pages /Count ${pages.length} /Kids [${planPages.map(item=>`${item.pageId} 0 R`).join(' ')}] >>`);fontEntries.forEach(([key,name],index)=>object(3+index,`<< /Type /Font /Subtype /Type1 /BaseFont /${name} /Encoding /WinAnsiEncoding >>`));for(const item of planPages){const fontDict=fontEntries.map(([key],index)=>`/F${key} ${3+index} 0 R`).join(' '),xobjects=item.page.images.map((image,index)=>`/${image.name} ${item.imageIds[index]} 0 R`).join(' '),stream=item.page.commands.join('\n'),streamBytes=new TextEncoder().encode(stream);object(item.pageId,`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${item.page.w} ${item.page.h}] /Resources << /Font << ${fontDict} >> ${xobjects?`/XObject << ${xobjects} >>`:''} >> /Contents ${item.contentId} 0 R >>`);object(item.contentId,`<< /Length ${streamBytes.length} >>\nstream\n${stream}\nendstream`);item.page.images.forEach((image,index)=>{const id=item.imageIds[index],alphaId=item.alphaIds[index],filter=image.filter?`/Filter /${image.filter} `:'';offsets[id]=length;add(`${id} 0 obj\n<< /Type /XObject /Subtype /Image /Width ${image.w} /Height ${image.h} /ColorSpace /DeviceRGB /BitsPerComponent 8 ${filter}${alphaId?`/SMask ${alphaId} 0 R `:''}/Length ${image.bytes.length} >>\nstream\n`);add(image.bytes);add('\nendstream\nendobj\n');if(alphaId){offsets[alphaId]=length;add(`${alphaId} 0 obj\n<< /Type /XObject /Subtype /Image /Width ${image.w} /Height ${image.h} /ColorSpace /DeviceGray /BitsPerComponent 8 /Length ${image.alphaBytes.length} >>\nstream\n`);add(image.alphaBytes);add('\nendstream\nendobj\n');}});}const xref=length,count=nextId-1;add(`xref\n0 ${count+1}\n0000000000 65535 f \n`);for(let id=1;id<=count;id+=1)add(`${String(offsets[id]).padStart(10,'0')} 00000 n \n`);add(`trailer\n<< /Size ${count+1} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF`);return new Blob(chunks,{type:'application/pdf'});}

const SGL5063_DOCX_MIME='application/vnd.openxmlformats-officedocument.wordprocessingml.document';
let SGL5063_CRC_TABLE=null;
function sgl5063Xml(value=''){return String(value??'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&apos;');}
function sgl5063Crc32(bytes){
  if(!SGL5063_CRC_TABLE){SGL5063_CRC_TABLE=new Uint32Array(256);for(let n=0;n<256;n+=1){let c=n;for(let k=0;k<8;k+=1)c=(c&1)?(0xedb88320^(c>>>1)):(c>>>1);SGL5063_CRC_TABLE[n]=c>>>0;}}
  let crc=0xffffffff;for(const byte of bytes)crc=SGL5063_CRC_TABLE[(crc^byte)&0xff]^(crc>>>8);return (crc^0xffffffff)>>>0;
}
function sgl5063DosStamp(date=new Date()){
  const year=Math.max(1980,date.getFullYear()),time=((date.getHours()&31)<<11)|((date.getMinutes()&63)<<5)|((Math.floor(date.getSeconds()/2))&31),day=((year-1980)<<9)|(((date.getMonth()+1)&15)<<5)|(date.getDate()&31);return {time,day};
}
function sgl5063ZipBlob(entries){
  const encoder=new TextEncoder(),stamp=sgl5063DosStamp(),locals=[],centrals=[];let offset=0;
  for(const entry of entries){const name=encoder.encode(String(entry.name||'')),data=typeof entry.data==='string'?encoder.encode(entry.data):(entry.data instanceof Uint8Array?entry.data:new Uint8Array(entry.data||0)),crc=sgl5063Crc32(data),local=new Uint8Array(30+name.length),lv=new DataView(local.buffer);lv.setUint32(0,0x04034b50,true);lv.setUint16(4,20,true);lv.setUint16(6,0x0800,true);lv.setUint16(8,0,true);lv.setUint16(10,stamp.time,true);lv.setUint16(12,stamp.day,true);lv.setUint32(14,crc,true);lv.setUint32(18,data.length,true);lv.setUint32(22,data.length,true);lv.setUint16(26,name.length,true);lv.setUint16(28,0,true);local.set(name,30);
    const central=new Uint8Array(46+name.length),cv=new DataView(central.buffer);cv.setUint32(0,0x02014b50,true);cv.setUint16(4,20,true);cv.setUint16(6,20,true);cv.setUint16(8,0x0800,true);cv.setUint16(10,0,true);cv.setUint16(12,stamp.time,true);cv.setUint16(14,stamp.day,true);cv.setUint32(16,crc,true);cv.setUint32(20,data.length,true);cv.setUint32(24,data.length,true);cv.setUint16(28,name.length,true);cv.setUint16(30,0,true);cv.setUint16(32,0,true);cv.setUint16(34,0,true);cv.setUint16(36,0,true);cv.setUint32(38,0,true);cv.setUint32(42,offset,true);central.set(name,46);locals.push(local,data);centrals.push(central);offset+=local.length+data.length;
  }
  const centralOffset=offset,centralSize=centrals.reduce((sum,item)=>sum+item.length,0),end=new Uint8Array(22),ev=new DataView(end.buffer);ev.setUint32(0,0x06054b50,true);ev.setUint16(4,0,true);ev.setUint16(6,0,true);ev.setUint16(8,centrals.length,true);ev.setUint16(10,centrals.length,true);ev.setUint32(12,centralSize,true);ev.setUint32(16,centralOffset,true);ev.setUint16(20,0,true);return new Blob([...locals,...centrals,end],{type:SGL5063_DOCX_MIME});
}
async function sgl5063PrepareDocxImage(source,{maxPx=1600,quality=.86}={}){
  if(!source)return null;const image=await loadImage(source),naturalW=Math.max(1,image.naturalWidth||image.width||1),naturalH=Math.max(1,image.naturalHeight||image.height||1),scale=Math.min(1,maxPx/Math.max(naturalW,naturalH)),width=Math.max(1,Math.round(naturalW*scale)),height=Math.max(1,Math.round(naturalH*scale)),canvas=document.createElement('canvas');canvas.width=width;canvas.height=height;const ctx=canvas.getContext('2d',{alpha:false});ctx.fillStyle='#fff';ctx.fillRect(0,0,width,height);ctx.drawImage(image,0,0,width,height);const blob=await new Promise((resolve,reject)=>canvas.toBlob(item=>item?resolve(item):reject(new Error('No se pudo preparar una imagen para Word.')),'image/jpeg',quality));return {bytes:new Uint8Array(await blob.arrayBuffer()),width,height,ext:'jpg',contentType:'image/jpeg'};
}
function sgl5063ContainEmu(image,maxW,maxH=maxW){const ratio=Math.min(maxW/Math.max(1,image?.width||1),maxH/Math.max(1,image?.height||1)),cx=Math.max(1,Math.round((image?.width||1)*ratio)),cy=Math.max(1,Math.round((image?.height||1)*ratio));return {cx,cy};}
function sgl5063DrawingXml(relId,cx,cy,docPrId,name='Imagen'){
  const safeName=sgl5063Xml(name);return `<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="${Math.round(cx)}" cy="${Math.round(cy)}"/><wp:effectExtent l="0" t="0" r="0" b="0"/><wp:docPr id="${Math.max(1,Number(docPrId)||1)}" name="${safeName}"/><wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="0" name="${safeName}"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="${sgl5063Xml(relId)}"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="${Math.round(cx)}" cy="${Math.round(cy)}"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`;
}
function sgl5063FontName(value,fallback='Arial'){const raw=String(value||'').replace(/["']/g,'').trim();if(/times\s+new\s+roman/i.test(raw))return 'Times New Roman';if(/arial/i.test(raw))return 'Arial';return fallback;}
function sgl5063FormatFromNode(node,parent){
  const next={...(parent||{})},tag=node?.tagName,style=node?.style;
  if(['B','STRONG'].includes(tag))next.bold=true;if(['I','EM'].includes(tag))next.italic=true;if(tag==='U')next.underline=true;
  if(style){
    if(style.fontWeight)next.bold=style.fontWeight==='bold'||Number(style.fontWeight)>=600;
    if(style.fontStyle)next.italic=['italic','oblique'].includes(style.fontStyle);
    if(style.textDecorationLine||style.textDecoration)next.underline=String(style.textDecorationLine||style.textDecoration).includes('underline');
    if(style.fontFamily)next.font=sgl5063FontName(style.fontFamily,next.font||'Arial');
    if(style.fontSize){const n=parseFloat(style.fontSize);if(n>0)next.size=Math.max(8,Math.min(72,/px$/.test(style.fontSize)?n*.75:n));}
    if(sgl532Color(style.color))next.color=sgl532Color(style.color);
    if(sgl532Color(style.backgroundColor))next.highlight=sgl532Color(style.backgroundColor);
  }return next;
}
function sgl5063RunXml(text,format){if(typeof text==='string')text=text.replace(/\u200b/g,'');if(text===null||text===undefined||text==='')return '';const f=format||{},props=[];if(f.bold)props.push('<w:b/>');if(f.italic)props.push('<w:i/>');if(f.underline)props.push('<w:u w:val="single"/>');if(f.font)props.push(`<w:rFonts w:ascii="${sgl5063Xml(f.font)}" w:hAnsi="${sgl5063Xml(f.font)}" w:cs="${sgl5063Xml(f.font)}"/>`);if(f.size){const half=Math.max(2,Math.round(Number(f.size)*2));props.push(`<w:sz w:val="${half}"/><w:szCs w:val="${half}"/>`);}if(f.color)props.push(`<w:color w:val="${sgl5063Xml(f.color)}"/>`);if(f.highlight)props.push(`<w:shd w:val="clear" w:fill="${sgl5063Xml(f.highlight)}"/>`);const chunks=String(text).split(/\n/);return chunks.map((chunk,index)=>`${index?'<w:r><w:br/></w:r>':''}${chunk!==''?`<w:r>${props.length?`<w:rPr>${props.join('')}</w:rPr>`:''}<w:t${/^\s|\s$|\s{2}/.test(chunk)?' xml:space="preserve"':''}>${sgl5063Xml(chunk)}</w:t></w:r>`:''}`).join('');}
function sgl5063InlineRuns(node,format){if(!node)return '';if(node.nodeType===3)return sgl5063RunXml(node.nodeValue||'',format);if(node.nodeType!==1)return '';const tag=String(node.tagName||'').toUpperCase();if(tag==='BR')return '<w:r><w:br/></w:r>';if(tag==='IMG'||tag==='FIGURE')return '';const next=sgl5063FormatFromNode(node,format);return [...node.childNodes].map(child=>sgl5063InlineRuns(child,next)).join('');}
function sgl5063ParagraphPr(node,{align='left',after=120,line=312}={}){let value=String(node?.style?.textAlign||node?.getAttribute?.('align')||align||'left').toLowerCase();if(value==='justify')value='both';if(!['left','center','right','both'].includes(value))value='left';const parts=[`<w:jc w:val="${value}"/>`,`<w:spacing w:after="${Math.max(0,Math.round(after))}" w:line="${Math.max(120,Math.round(line))}" w:lineRule="auto"/>`],margin=String(node?.style?.marginLeft||'');if(margin){const n=parseFloat(margin);if(Number.isFinite(n)&&n>0){const twips=Math.round((/px$/i.test(margin)?n*.75:n)*20);parts.push(`<w:ind w:left="${Math.min(5000,twips)}"/>`);}}return `<w:pPr>${parts.join('')}</w:pPr>`;}
function sgl5063ParagraphXml(node,format,options={}){const prefix=options.prefix?sgl5063RunXml(options.prefix,{...format,bold:!!options.prefixBold}):'',runs=prefix+sgl5063InlineRuns(node,format);return `<w:p>${sgl5063ParagraphPr(node,options)}${runs||'<w:r><w:t></w:t></w:r>'}</w:p>`;}
function sgl5063SimpleParagraph(text,{font='Arial',size=11,bold=false,color='',align='left',after=80,upper=false}={}){const clean=upper?String(text||'').toUpperCase():String(text||'');return `<w:p><w:pPr><w:jc w:val="${align==='justify'?'both':align}"/><w:spacing w:after="${after}"/></w:pPr>${sgl5063RunXml(clean,{font,size,bold,color})}</w:p>`;}
function sgl5063TableCell(content,width,{align='center',vAlign='center'}={}){return `<w:tc><w:tcPr><w:tcW w:w="${Math.round(width)}" w:type="dxa"/><w:vAlign w:val="${vAlign}"/></w:tcPr><w:p><w:pPr><w:jc w:val="${align}"/><w:spacing w:after="0"/></w:pPr>${content||'<w:r><w:t></w:t></w:r>'}</w:p></w:tc>`;}
function sgl5063HeaderXml({official,clubLogo,mission,title,identity,contentTwips}){
  const left=940,right=940,center=Math.max(1000,contentTwips-left-right),maxLogoEmu=43*12700,leftExtent=official?sgl5063ContainEmu(official,maxLogoEmu,maxLogoEmu):null,rightExtent=clubLogo?sgl5063ContainEmu(clubLogo,maxLogoEmu,maxLogoEmu):null,leftRun=leftExtent?sgl5063DrawingXml('rIdLogoOfficial',leftExtent.cx,leftExtent.cy,1,'Logo institucional'):'',rightRun=rightExtent?sgl5063DrawingXml('rIdLogoClub',rightExtent.cx,rightExtent.cy,2,'Logo del club'):'',centerContent=`${mission?sgl5063SimpleParagraph(mission,{font:'Arial',size:8.5,bold:true,align:'center',after:10,upper:true}):''}${sgl5063SimpleParagraph(title,{font:'Arial',size:13.5,bold:true,color:'1F739E',align:'center',after:10})}${sgl5063SimpleParagraph(identity,{font:'Arial',size:10.5,align:'center',after:0})}`;
  return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:tbl><w:tblPr><w:tblW w:w="${contentTwips}" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="75" w:type="dxa"/><w:left w:w="75" w:type="dxa"/><w:bottom w:w="75" w:type="dxa"/><w:right w:w="75" w:type="dxa"/></w:tblCellMar><w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="${left}"/><w:gridCol w:w="${center}"/><w:gridCol w:w="${right}"/></w:tblGrid><w:tr>${sgl5063TableCell(leftRun,left)}<w:tc><w:tcPr><w:tcW w:w="${center}" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr>${centerContent}</w:tc>${sgl5063TableCell(rightRun,right)}</w:tr></w:tbl><w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="12" w:space="4" w:color="1F739E"/></w:pBdr><w:spacing w:after="0"/></w:pPr></w:p></w:hdr>`;
}
function sgl5063FooterXml(data,contentTwips){const info=reportInstitutionInfo(data),left=Math.floor(contentTwips/3),middle=Math.floor(contentTwips/3),right=contentTwips-left-middle,cell=(text,width,align)=>`<w:tc><w:tcPr><w:tcW w:w="${width}" w:type="dxa"/></w:tcPr>${sgl5063SimpleParagraph(text,{font:'Arial',size:8.5,align,after:0,upper:true})}</w:tc>`;return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:tbl><w:tblPr><w:tblW w:w="${contentTwips}" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="75" w:type="dxa"/><w:left w:w="75" w:type="dxa"/><w:bottom w:w="75" w:type="dxa"/><w:right w:w="75" w:type="dxa"/></w:tblCellMar><w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="${left}"/><w:gridCol w:w="${middle}"/><w:gridCol w:w="${right}"/></w:tblGrid><w:tr>${cell(`IGLESIA: ${info.church}`,left,'left')}${cell(`DISTRITO: ${info.district}`,middle,'center')}${cell(`REGIÓN: ${info.region}`,right,'right')}</w:tr></w:tbl></w:ftr>`;}
async function sgl5063BuildDocx(data){
  const font=data.aspirant.reportFont==='Times New Roman'?'Times New Roman':'Arial',fontSize=Math.max(10,Math.min(14,Number(data.submission?.reportFontSize||12)||12)),orientation=(data.submission?.orientation==='landscape'?'landscape':'portrait'),pageW=orientation==='landscape'?15840:12240,pageH=orientation==='landscape'?12240:15840,marginLeft=1701,marginRight=1134,marginTop=1390,marginBottom=850,contentTwips=pageW-marginLeft-marginRight,entries=[],docRels=[],headerRels=[],used=new Set(),bodyImages=new Map();let relCounter=5,mediaCounter=1,docPrCounter=20;
  // 5.0.64: los logos se remuestrean a 192 px; Word recibe un tamaño físico fijo,
  // por lo que nunca usa la resolución original (p.ej. 2048x2048) para dibujarlos.
  const official=await sgl5063PrepareDocxImage(data.officialLogo,{maxPx:192,quality:.82}),clubLogo=data.club?.logo?await sgl5063PrepareDocxImage(data.club.logo,{maxPx:192,quality:.82}):null;
  if(official){entries.push({name:'word/media/header-official.jpg',data:official.bytes});headerRels.push('<Relationship Id="rIdLogoOfficial" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/header-official.jpg"/>');}
  if(clubLogo){entries.push({name:'word/media/header-club.jpg',data:clubLogo.bytes});headerRels.push('<Relationship Id="rIdLogoClub" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/header-club.jpg"/>');}
  const attachedById=new Map((data.attached||[]).map(file=>[String(file.id),file]));
  async function imageFor(file){if(!file?.dataUrl)return null;const key=String(file.id||file.dataUrl.slice(0,64));if(bodyImages.has(key))return bodyImages.get(key);const image=await sgl5063PrepareDocxImage(file.dataUrl,{maxPx:1800,quality:.86}),relId=`rIdImage${relCounter++}`,target=`media/report-${mediaCounter++}.jpg`;entries.push({name:`word/${target}`,data:image.bytes});docRels.push(`<Relationship Id="${relId}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="${target}"/>`);const item={...image,relId};bodyImages.set(key,item);return item;}
  async function imageParagraph(file,figure=null){const image=await imageFor(file);if(!image)return '';used.add(String(file.id));const box=data.reportPhotoBoxes?.[String(file.id)]||data.submission?.reportPhotoBoxes?.[String(file.id)]||{},scale=Math.max(.01,Math.min(1,Number(box.scale||figure?.dataset?.sglScale||file.reportScale||70)/100)),maxWidthIn=(contentTwips/1440)*scale,maxHeightIn=5.2,ratio=Math.min((maxWidthIn*914400)/Math.max(1,image.width),(maxHeightIn*914400)/Math.max(1,image.height)),cx=Math.max(1,Math.round(image.width*ratio)),cy=Math.max(1,Math.round(image.height*ratio)),align=String(figure?.dataset?.sglAlign||'center'),mode=box.layout||figure?.dataset?.sglLayout||'inline',anchor=box.anchor||figure?.dataset?.sglAnchor||'page';let drawing=sgl5063DrawingXml(image.relId,cx,cy,docPrCounter++,`Fotografía ${sgl5063Xml(file.name||file.id||'')}`).replace('<a:xfrm>',`<a:xfrm rot="${Math.round((Number(box.angle||figure?.dataset?.sglAngle)||0)*60000)}">`);
    if(mode!=='inline'){const floating=['front','behind'].includes(mode),left=Math.round((Number(box.left)||0)*9525),top=Math.round((Number(box.top)||0)*9525),wrap=floating?'<wp:wrapNone/>':mode==='topbottom'?'<wp:wrapTopAndBottom/>':mode==='tight'||mode==='through'?`<wp:wrap${mode==='through'?'Through':'Tight'} wrapText="bothSides"><wp:wrapPolygon edited="0"><wp:start x="0" y="0"/><wp:lineTo x="21600" y="0"/><wp:lineTo x="21600" y="21600"/><wp:lineTo x="0" y="21600"/><wp:lineTo x="0" y="0"/></wp:wrapPolygon></wp:wrap${mode==='through'?'Through':'Tight'}>`:'<wp:wrapSquare wrapText="bothSides"/>';drawing=drawing.replace(/<wp:inline[^>]*>/,`<wp:anchor distT="0" distB="0" distL="0" distR="0" simplePos="0" relativeHeight="${mode==='behind'?0:251658240}" behindDoc="${mode==='behind'?1:0}" locked="0" layoutInCell="1" allowOverlap="1"><wp:simplePos x="0" y="0"/><wp:positionH relativeFrom="column"><wp:posOffset>${left}</wp:posOffset></wp:positionH><wp:positionV relativeFrom="${anchor==='text'?'paragraph':'page'}"><wp:posOffset>${anchor==='text'?Math.round((Number(box.anchorOffset)||0)*9525):top+marginTop*635}</wp:posOffset></wp:positionV>`).replace('<wp:docPr',wrap+'<wp:docPr').replace('</wp:inline>','</wp:anchor>');}
    return `<w:p><w:pPr><w:jc w:val="${['left','right','center'].includes(align)?align:'center'}"/><w:spacing w:after="180"/></w:pPr>${drawing}</w:p>`;}
  const mark=data.submission?.reportWatermark,markFile=attachedById.get(mark?.fileId);let watermarkImage=null;
  if(markFile?.dataUrl&&/^data:image\/png;base64,/.test(markFile.dataUrl)){const bytes=Uint8Array.from(atob(markFile.dataUrl.split(',')[1]),c=>c.charCodeAt(0)),image=await loadImage(markFile.dataUrl),relId='rIdWatermark548';watermarkImage={width:image.naturalWidth,height:image.naturalHeight,relId};entries.push({name:'word/media/watermark548.png',data:bytes});docRels.push('<Relationship Id="'+relId+'" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/watermark548.png"/>');used.add(String(mark.fileId));}
  function watermarkDrawing(pageIndex){if(!watermarkImage||(!mark.allPages&&Number(mark.page)!==pageIndex+1))return '';const image=watermarkImage,ratio=Math.min(pageW*635*(Math.max(15,Math.min(90,Number(mark.size)||60))/100)/image.width,pageH*635*.9/image.height),cx=Math.round(image.width*ratio),cy=Math.round(image.height*ratio),xml=sgl5063DrawingXml(image.relId,cx,cy,docPrCounter++,'Marca de agua');return '<w:p>'+xml.replace(/<wp:inline[^>]*>/,'<wp:anchor distT="0" distB="0" distL="0" distR="0" simplePos="0" relativeHeight="0" behindDoc="1" locked="0" layoutInCell="1" allowOverlap="1"><wp:simplePos x="0" y="0"/><wp:positionH relativeFrom="page"><wp:align>center</wp:align></wp:positionH><wp:positionV relativeFrom="page"><wp:align>center</wp:align></wp:positionV>').replace('<wp:docPr','<wp:wrapNone/><wp:docPr').replace('</wp:inline>','</wp:anchor>').replace(/<a:blip r:embed="([^"]+)"\/>/,'<a:blip r:embed="$1"><a:alphaModFix amt="'+Math.round(Math.max(.05,Math.min(.5,Number(mark.opacity)||.15))*100000)+'"/></a:blip>')+'</w:p>';}
  async function htmlBlocks(source,inCell=false){const parser=new DOMParser(),doc=parser.parseFromString(`<body>${String(source||'')}</body>`,'text/html'),root=doc.body,blocks=[];const base={font,size:fontSize};async function walk(node,listDepth=0){if(node.nodeType===3){if(String(node.nodeValue||'').trim())blocks.push(`<w:p>${sgl5063ParagraphPr(null,{})}${sgl5063RunXml(node.nodeValue,base)}</w:p>`);return;}if(node.nodeType!==1)return;const tag=String(node.tagName||'').toUpperCase();if(tag==='TABLE'){
 const rows=[...node.rows],pending=[],cols=[...node.querySelectorAll(':scope > colgroup > col')],count=Math.max(1,cols.length,...rows.map(row=>[...row.cells].reduce((n,c)=>n+c.colSpan,0))),weights=Array.from({length:count},(_,i)=>parseFloat(cols[i]?.style.width)||1),total=weights.reduce((a,b)=>a+b,0),widths=weights.map(w=>Math.round(contentTwips*w/total)),cellWidth=(col,span)=>widths.slice(col,col+span).reduce((a,b)=>a+b,0);let xml='<w:tbl><w:tblPr><w:tblW w:w="5000" w:type="pct"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="75" w:type="dxa"/><w:left w:w="75" w:type="dxa"/><w:bottom w:w="75" w:type="dxa"/><w:right w:w="75" w:type="dxa"/></w:tblCellMar><w:tblBorders>'+['top','left','bottom','right','insideH','insideV'].map(side=>'<w:'+side+' w:val="single" w:sz="4" w:color="555555"/>').join('')+'</w:tblBorders></w:tblPr><w:tblGrid>'+widths.map(w=>'<w:gridCol w:w="'+w+'"/>').join('')+'</w:tblGrid>';
 for(const row of rows){const rowHeight=parseFloat(row.style.height)||0;xml+='<w:tr>'+(rowHeight?'<w:trPr><w:trHeight w:val="'+Math.round(rowHeight*15)+'" w:hRule="atLeast"/></w:trPr>':'');let col=0;const occupied=()=>{while(pending[col]?.left>0){const p=pending[col];xml+='<w:tc><w:tcPr><w:tcW w:w="'+cellWidth(col,p.span)+'" w:type="dxa"/>'+(p.span>1?'<w:gridSpan w:val="'+p.span+'"/>':'')+'<w:vMerge/></w:tcPr><w:p/></w:tc>';p.left--;col+=p.span;}};
 for(const cell of row.cells){occupied();const span=cell.colSpan||1,height=cell.rowSpan||1;xml+='<w:tc><w:tcPr><w:tcW w:w="'+cellWidth(col,span)+'" w:type="dxa"/>'+(span>1?'<w:gridSpan w:val="'+span+'"/>':'')+(height>1?'<w:vMerge w:val="restart"/>':'')+'</w:tcPr>'+((await htmlBlocks(cell.innerHTML,true)).join('')||'<w:p/>')+'</w:tc>';if(height>1)pending[col]={left:height-1,span};col+=span;}
 occupied();xml+='</w:tr>';}
 blocks.push(xml+'</w:tbl>');return;
 }if(tag==='FIGURE'){const id=String(node.dataset?.sglPhotoId||node.getAttribute('data-sgl-photo-id')||''),file=attachedById.get(id);if(file)blocks.push(await imageParagraph(file,node));return;}if(tag==='IMG'){const src=node.getAttribute('src')||'';if(src.startsWith('data:image/')){const pseudo={id:`inline-${mediaCounter}`,dataUrl:src,name:'Imagen'};blocks.push(await imageParagraph(pseudo,node));}return;}if(tag==='UL'||tag==='OL'){let index=1;for(const child of [...node.children])if(String(child.tagName||'').toUpperCase()==='LI'){const prefix=tag==='UL'?(node.style.listStyleType==='circle'?'○ ':node.style.listStyleType==='square'?'▪ ':'• '):`${index++}. `;blocks.push(sgl5063ParagraphXml(child,base,{prefix,after:80,line:312}));}return;}if(tag==='P'||tag==='DIV'||tag==='LI'||/^H[1-6]$/.test(tag)){blocks.push(sgl5063ParagraphXml(node,base,{after:inCell?60:120,line:312}));return;}for(const child of [...node.childNodes])await walk(child,listDepth);}
    for(const child of [...root.childNodes])await walk(child,0);return blocks;
  }
  const body=[];body.push(sgl5063SimpleParagraph(data.requirement?.areaTitle||'',{font:'Arial',size:10.5,bold:true,color:'1F739E',after:80,upper:true}));body.push(sgl5063SimpleParagraph(data.requirement?.official_text||'',{font,size:11.5,bold:true,after:80}));body.push(sgl5063SimpleParagraph(reportRequirementFillDate(data),{font:'Arial',size:10.5,bold:true,color:'1F739E',align:'right',after:160}));for(const selection of data.submission?.selections||[])body.push(sgl5063SimpleParagraph(selection,{font,size:fontSize,bold:true,after:120}));
  const layoutPages=Array.isArray(data.reportLayoutPages)&&data.reportLayoutPages.length?data.reportLayoutPages:(Array.isArray(data.submission?.reportLayoutPages)?data.submission.reportLayoutPages:[]);if(layoutPages.length){for(let pageIndex=0;pageIndex<layoutPages.length;pageIndex+=1){if(pageIndex>0)body.push('<w:p><w:r><w:br w:type="page"/></w:r></w:p>');body.push(watermarkDrawing(pageIndex));body.push(...await htmlBlocks(layoutPages[pageIndex]||''));}}else {body.push(watermarkDrawing(0));body.push(...await htmlBlocks(reportHtmlFromSubmission(data.submission)));}
  for(const file of data.attached||[])if(file.kind==='photo'&&!used.has(String(file.id)))body.push(await imageParagraph(file));
  const sect=`<w:sectPr><w:headerReference w:type="default" r:id="rIdHeader"/><w:footerReference w:type="default" r:id="rIdFooter"/><w:pgSz w:w="${pageW}" w:h="${pageH}"${orientation==='landscape'?' w:orient="landscape"':''}/><w:pgMar w:top="${marginTop}" w:right="${marginRight}" w:bottom="${marginBottom}" w:left="${marginLeft}" w:header="320" w:footer="320" w:gutter="0"/><w:cols w:space="720"/><w:docGrid w:linePitch="360"/></w:sectPr>`;
  const documentXml=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:body>${body.join('')}${sect}</w:body></w:document>`;
  const styles=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="${sgl5063Xml(font)}" w:hAnsi="${sgl5063Xml(font)}" w:cs="${sgl5063Xml(font)}"/><w:sz w:val="${fontSize*2}"/><w:szCs w:val="${fontSize*2}"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="312" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:qFormat/></w:style></w:styles>`;
  const settings=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:zoom w:percent="100"/><w:defaultTabStop w:val="720"/><w:compat><w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/></w:compat></w:settings>`;
  const headerXml=sgl5063HeaderXml({official,clubLogo,mission:data.mission||'',title:data.reportTitle||'INFORME CLASE AGRUPADA',identity:data.identity||'',contentTwips}),footerXml=sgl5063FooterXml(data,contentTwips),documentRels=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rIdStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rIdSettings" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/><Relationship Id="rIdHeader" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/><Relationship Id="rIdFooter" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>${docRels.join('')}</Relationships>`,headerRelXml=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">${headerRels.join('')}</Relationships>`,now=new Date().toISOString(),core=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>${sgl5063Xml(reportBaseName(data.requirement||{}))}</dc:title><dc:creator>Sistema de Gestión de Música</dc:creator><cp:lastModifiedBy>Sistema de Gestión de Música 5.48</cp:lastModifiedBy><dcterms:created xsi:type="dcterms:W3CDTF">${now}</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">${now}</dcterms:modified></cp:coreProperties>`,app=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><Application>Sistema de Gestión de Música</Application><AppVersion>6.0</AppVersion></Properties>`,contentTypes=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="jpg" ContentType="image/jpeg"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/><Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/><Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`,rootRels=`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`;
  entries.push({name:'[Content_Types].xml',data:contentTypes},{name:'_rels/.rels',data:rootRels},{name:'docProps/core.xml',data:core},{name:'docProps/app.xml',data:app},{name:'word/document.xml',data:documentXml},{name:'word/styles.xml',data:styles},{name:'word/settings.xml',data:settings},{name:'word/header1.xml',data:headerXml},{name:'word/footer1.xml',data:footerXml},{name:'word/_rels/document.xml.rels',data:documentRels},{name:'word/_rels/header1.xml.rels',data:headerRelXml});return sgl5063ZipBlob(entries);
}

async function exportWord(requirementId) {
  try{
    toast('Generando Word moderno…');
    const data=await printableRequirement(requirementId),blob=await sgl5063BuildDocx(data),result=await saveBlobWithDialog(reportFileName(data,'docx'),blob,'Documento de Word moderno|*.docx');
    if(result?.ok)toast(`Documento DOCX guardado en ${result.path}.`);
  }catch(error){await logClientError('Exportación Word DOCX',error);toast(`No se pudo generar el DOCX: ${error.message}`);}
}

function renderDrive() { return false; }

async function googleLogin() { return false; }

async function refreshDriveStatus(){ return false; }
async function saveDriveFolder(){ return ''; }
async function disconnectDrive() { return false; }
async function sendReportToDrive(requirementId){ return false; }

function scopedDatabase(user=currentUser()) {
  if (user.role==='admin') return structuredClone(database);
  const aspirantIds=accessibleAspirants(user).map(item=>item.id), userIds=new Set([user.id,...aspirantIds]);
  return {...structuredClone(database),users:database.users.filter(item=>userIds.has(item.id)).map(item=>({...item,password:undefined,drive:undefined})),clubs:database.clubs.filter(club=>database.users.some(item=>userIds.has(item.id)&&item.clubId===club.id)),materials:structuredClone(accessibleMaterials(user)),progress:Object.fromEntries(Object.entries(database.progress).filter(([id])=>aspirantIds.includes(id))),configs:Object.fromEntries(Object.entries(database.configs).filter(([key])=>aspirantIds.some(id=>key.startsWith(`${id}:`)))),audit:[],settings:undefined,conflicts:[]};
}

function renderBackup() {
  const user=currentUser(); setTop('Guardar una copia',ROLE_LABELS[user.rawRole624||user.role]);
  const scope=user.role==='admin'?'todos los usuarios, clubes y expedientes':['aspirant','pastor_aspirant'].includes(user.role)?'tu expediente, informes y archivos':'los expedientes de tus Aspirantes asignados';
  setContent(`<div class="page-head"><div><div class="eyebrow">Copia según permisos</div><h1 class="page-title">Respaldos</h1><p class="subtitle">Tu copia incluye ${scope}.</p></div></div><div class="backup-options"><div class="backup-card"><h4>Copia protegida</h4><p>Archivo cifrado con contraseña para guardar en PC, USB o disco externo.</p><button class="btn primary block" onclick="exportEncryptedBackup()">Crear copia</button></div><div class="backup-card"><h4>Copia JSON legible</h4><p>Datos estructurados para migración o revisión técnica. No incluye contraseñas ajenas.</p><button class="btn ghost block" onclick="exportJsonBackup()">Descargar JSON</button></div><div class="backup-card"><h4>Google Drive</h4><p>Guarda la copia dentro de la carpeta personal o administrativa habilitada.</p><button class="btn ghost block" onclick="go('drive')">Abrir Drive</button></div></div><div class="card" style="margin-top:16px"><h3>Restaurar copia protegida</h3><p class="subtitle">Solo el Administrador puede restaurar una copia completa. La operación crea un punto de seguridad antes de reemplazar datos.</p><label class="btn ghost ${user.role==='admin'?'':'hidden'}">Elegir copia<input hidden type="file" accept=".cabackup" onchange="restoreEncryptedBackup(this.files[0]);this.value=''"/></label></div>`);
}

async function backupPayload() {
  const scoped=scopedDatabase(), allowedIds=new Set(scoped.users.map(item=>item.id)), files=(await allFiles()).filter(file=>allowedIds.has(file.ownerId));
  if (currentUser().role!=='admin') scoped.users=scoped.users.map(item=>({...item,password:undefined}));
  return {format:'MinisterioMusicaBackup',version:APP_VERSION,exportedAt:isoNow(),exportedBy:{id:currentUser().id,name:currentUser().fullName,role:currentUser().role},database:scoped,files};
}

async function exportJsonBackup() { const payload=await backupPayload(),name=`MinisterioMusica_${currentUser().role}_${localDate()}.json`,result=await saveBlobWithDialog(name,new Blob([JSON.stringify(payload,null,2)],{type:'application/json'}),'Copia JSON|*.json');if(!result?.ok)return;addAudit('Copia JSON exportada','backup',currentUser().role); await persist(false);toast(`Copia guardada en ${result.path}.`); }

async function exportEncryptedBackup() {
  const password=prompt('Escribe una contraseña para proteger la copia (mínimo 8 caracteres):',''); if (!password||password.length<8) return toast('La copia requiere una contraseña de al menos 8 caracteres.');
  const salt=crypto.getRandomValues(new Uint8Array(16)), iv=crypto.getRandomValues(new Uint8Array(12)), raw=await deriveKeyBytes(password,salt,180000,32), key=await crypto.subtle.importKey('raw',raw,{name:'AES-GCM'},false,['encrypt']), payload=await backupPayload(), encrypted=await crypto.subtle.encrypt({name:'AES-GCM',iv},key,new TextEncoder().encode(JSON.stringify(payload))), envelope={format:'MinisterioMusicaEncryptedBackup',version:1,kdf:'PBKDF2-SHA256',iterations:180000,salt:toBase64(salt),iv:toBase64(iv),ciphertext:toBase64(new Uint8Array(encrypted))};
  const result=await saveBlobWithDialog(`MinisterioMusica_Copia_${currentUser().role}_${localDate()}.cabackup`,new Blob([JSON.stringify(envelope)],{type:'application/octet-stream'}),'Copia protegida|*.cabackup');if(!result?.ok)return;addAudit('Copia protegida exportada','backup',currentUser().role); await persist(false); toast(`Copia protegida guardada en ${result.path}.`);
}

async function restoreEncryptedBackup(file) {
  if (!file||currentUser().role!=='admin') return;
  const password=prompt('Contraseña de la copia:','');
  if (!password) return;
  try {
    const envelope=JSON.parse(await file.text());
    const salt=fromBase64(envelope.salt),iv=fromBase64(envelope.iv);
    const raw=await deriveKeyBytes(password,salt,envelope.iterations||180000,32);
    const key=await crypto.subtle.importKey('raw',raw,{name:'AES-GCM'},false,['decrypt']);
    const plain=await crypto.subtle.decrypt({name:'AES-GCM',iv},key,fromBase64(envelope.ciphertext));
    const payload=JSON.parse(new TextDecoder().decode(plain));
    if(!['MinisterioMusicaBackup','ClaseAgrupadaBackup'].includes(payload.format))throw new Error('Formato no válido');
    if(payload.format==='ClaseAgrupadaBackup')payload.format='MinisterioMusicaBackup';
    const safety=await backupPayload();
    await kvSet(`safety_${Date.now()}`,safety);

    if(typeof window.sgl5028PublishAdminRestore==='function'){
      const published=await window.sgl5028PublishAdminRestore(payload);
      if(published)return;
    }

    const previousGeneration=Math.max(1,Number(database?.restoreGeneration||1)||1);
    database=payload.database;
    migrateDatabase();
    database.restoreGeneration=Math.max(previousGeneration+1,Number(database.restoreGeneration||1)||1);
    database.restoreId=`local_restore_${Date.now()}`;
    database.restoredAt=isoNow();
    database.restorePending=false;
    database.syncQueue=[];
    database.syncHistory=[];
    database.messageOutbox=[];
    database.messageCache=[];
    await clearFiles();
    for(const item of payload.files||[])await filePut(item);
    sessionStorage.removeItem('musica_session_user');
    localStorage.removeItem('musica_session_user');
    await persist(false);
    if(!REMOTE_WEB)await writeNativeSnapshot();
    renderLogin('El sistema fue restaurado. Inicia sesión nuevamente con la copia restaurada.');
  } catch(error){
    console.error('SGL 5.0.52 restauración',error);
    toast(error?.message||'No se pudo restaurar: contraseña, archivo o servidor incorrecto.');
  }
}

function renderProfile() {
  const user=currentUser(), club=clubFor(user),driveLinked=!!user.drive?.connected; setTop('Mi perfil',ROLE_LABELS[user.rawRole624||user.role]);
  setContent(`<div class="page-head"><div><div class="eyebrow">Cuenta personal</div><h1 class="page-title">Perfil</h1></div></div><div class="card profile"><div><div class="profile-photo" id="profile_preview">${avatarMarkup(user)}</div><button class="btn ghost small block" style="margin-top:8px" onclick="openProfilePhotoStudio()">Estudio de foto</button></div><div><h3>${esc(user.fullName)}</h3><p class="subtitle">${ROLE_LABELS[user.rawRole624||user.role]} · ${esc(club?.name||'Sin Club')}</p><div class="field-row"><label class="field-group">Correo<input id="profile_email" type="email" class="field" autocomplete="email" value="${esc(user.email||'')}"></label><label class="field-group">Usuario<input class="field" value="${esc(user.username)}" readonly></label><label class="field-group">Fecha de nacimiento<input class="field" value="${esc(formatDate(user.birthDate))}" readonly></label><label class="field-group">Teléfono celular<input class="field" value="${esc(user.phone||'No registrado')}" readonly></label><label class="field-group">Letra de mis informes<select id="profile_font" class="field" onchange="saveOwnReportFont(this.value)"><option ${user.reportFont!=='Times New Roman'?'selected':''}>Arial</option><option ${user.reportFont==='Times New Roman'?'selected':''}>Times New Roman</option></select></label><label class="field-group">Tema de color<select id="profile_theme" class="field" onchange="saveOwnTheme(this.value)">${Object.entries(THEME_LABELS).map(([id,label])=>`<option value="${id}" ${user.theme===id?'selected':''}>${esc(label)}</option>`).join('')}</select></label></div>${driveLinked?'<div class="notice warn" style="margin-top:12px"><b>Google Drive está vinculado.</b> Si cambias el correo, la vinculación o el envío de archivos a Drive podría requerir una nueva validación. Después del cambio verifica el estado de Google Drive.</div>':''}<div class="actions"><button class="btn primary" onclick="saveOwnEmail()">Guardar correo</button><button class="btn ghost" onclick="changeOwnPassword()">Cambiar mi contraseña</button><button class="btn ghost" onclick="openProfilePhotoStudio()">Editar foto y filtros</button></div></div></div>`);
}

async function saveOwnReportFont(font){const user=currentUser();user.reportFont=font==='Times New Roman'?'Times New Roman':'Arial';await persist(true,{action:'Letra de informes actualizada',entity:user.id,detail:user.reportFont});toast('Preferencia de letra guardada.');}
async function saveOwnTheme(theme){const user=currentUser();if(!user)return;user.theme=THEME_LABELS[theme]?theme:'classic_blue';applyTheme();await persist(true,{action:'Tema personal actualizado',entity:user.id,detail:THEME_LABELS[user.theme]});toast('Tema personal guardado.');}
async function saveOwnEmail(){
  const user=currentUser();if(!user)return;const field=document.getElementById('profile_email'),next=String(field?.value||'').trim().toLowerCase(),current=String(user.email||'').trim().toLowerCase();
  if(!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(next))return toast('Escribe un correo válido.');
  if(next===current)return toast('El correo no cambió.');
  const duplicate=(database.users||[]).find(item=>item.id!==user.id&&String(item.email||'').trim().toLowerCase()===next);if(duplicate)return toast('Ese correo ya está registrado en otra cuenta.');
  if(user.drive?.connected&&!confirm('Tu cuenta tiene Google Drive vinculado. Si cambias el correo, la vinculación o el envío de archivos puede requerir una nueva validación. ¿Deseas continuar?')){if(field)field.value=user.email||'';return;}
  const previous=user.email;user.email=next;user.updatedAt=isoNow();
  try{await persist(true,{action:'Correo personal actualizado',entity:user.id,detail:`${previous||''} → ${next}`});renderProfile();toast('Correo actualizado. Verifica Google Drive si estaba vinculado.');}
  catch(error){user.email=previous;if(field)field.value=previous||'';toast(`No se pudo actualizar el correo: ${error.message}`);}
}

function openProfilePhotoStudio(target=null) {
  const me=currentUser();photoStudioTarget=target||{type:'user',userId:me?.id,returnPage:'profile'};
  const targetUser=photoStudioTarget?.type==='user'?userById(photoStudioTarget.userId):null,targetClass=photoStudioTarget?.type==='member'?accessibleClasses().find(item=>item.id===photoStudioTarget.classId):null,targetMember=targetClass?classWork(targetClass).members.find(member=>member.id===photoStudioTarget.memberId):null,targetName=photoStudioTarget?.type==='club'?photoStudioTarget.name:targetMember?.fullName||targetUser?.fullName||'perfil',source=photoStudioTarget?.type==='club'?photoStudioTarget.source:targetMember?.photo||targetUser?.photo||'';
  profileStudioImage=null;profileStudioRotation=0;profileStudioOffset={x:0,y:0};profileStudioDrag=null;
  showModal(`<div class="modal"><div class="modal-head"><div><h3>Estudio de foto · ${esc(targetName)}</h3><p>Arrastra la foto con la mano para centrar el recorte y usa Zoom para acercarla o reducirla.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body photo-studio"><div class="photo-stage"><canvas id="profile_canvas" width="600" height="600" aria-label="Vista previa: arrastra para cambiar la posición"></canvas><div class="photo-stage-note">✋ Arrastra la foto para ubicarla</div></div><div class="photo-controls"><label class="field-group">Elegir foto<input id="profile_file" type="file" class="field" accept="image/*" capture="user" onchange="loadProfileStudioFile(this.files[0])"></label><div class="section-label">Filtros rápidos</div><div class="photo-presets">${[['natural','Natural'],['vivid','Vivo'],['soft','Suave'],['warm','Cálido'],['cool','Frío'],['portrait','Retrato'],['contrast','Contraste'],['vintage','Vintage'],['bw','Blanco y negro'],['sepia','Sepia']].map(([key,label])=>`<button type="button" onclick="applyPhotoPreset('${key}')">${label}</button>`).join('')}</div><div class="photo-sliders"><label class="field-group">Zoom <output id="photo_zoom_value">100%</output><input id="photo_zoom" type="range" min="25" max="220" value="100" oninput="renderProfileStudio()"></label><label class="field-group">Brillo <output id="photo_brightness_value">100%</output><input id="photo_brightness" type="range" min="60" max="140" value="100" oninput="renderProfileStudio()"></label><label class="field-group">Contraste <output id="photo_contrast_value">100%</output><input id="photo_contrast" type="range" min="60" max="160" value="100" oninput="renderProfileStudio()"></label><label class="field-group">Saturación <output id="photo_saturation_value">100%</output><input id="photo_saturation" type="range" min="0" max="180" value="100" oninput="renderProfileStudio()"></label><label class="field-group">Blanco y negro <output id="photo_gray_value">0%</output><input id="photo_gray" type="range" min="0" max="100" value="0" oninput="renderProfileStudio()"></label><label class="field-group">Sepia <output id="photo_sepia_value">0%</output><input id="photo_sepia" type="range" min="0" max="100" value="0" oninput="renderProfileStudio()"></label><label class="field-group">Tono <output id="photo_hue_value">0°</output><input id="photo_hue" type="range" min="-30" max="30" value="0" oninput="renderProfileStudio()"></label><label class="field-group">Suavizado <output id="photo_blur_value">0</output><input id="photo_blur" type="range" min="0" max="3" step="0.25" value="0" oninput="renderProfileStudio()"></label><label class="field-group">Viñeta <output id="photo_vignette_value">0%</output><input id="photo_vignette" type="range" min="0" max="70" value="0" oninput="renderProfileStudio()"></label></div><div class="actions"><button type="button" class="btn ghost" onclick="rotateProfileStudio(-90)">↶ Girar</button><button type="button" class="btn ghost" onclick="rotateProfileStudio(90)">Girar ↷</button><button type="button" class="btn ghost" onclick="resetProfileStudio()">Restablecer</button></div><div class="notice">Se guarda exactamente el encuadre visible, cuadrado y en alta calidad.</div></div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cancelar</button><button id="save_profile_photo" class="btn primary" onclick="saveProfileStudio()" disabled>Guardar foto</button></div></div>`);
  bindProfileStudioDrag();
  if(source)loadProfileStudioSource(source);
}
function openClassMemberPhotoStudio(classId,memberId){const item=accessibleClasses().find(entry=>entry.id===classId),member=item?classWork(item).members.find(entry=>entry.id===memberId):null;if(!item||!member)return toast('No se encontró el Conquistador.');openProfilePhotoStudio({type:'member',classId,memberId,returnPage:'classes'});}
function openUserPhotoStudio(userId){if(currentUser()?.role!=='admin')return;const user=userById(userId);if(!user)return toast('No se encontró el usuario.');openProfilePhotoStudio({type:'user',userId,returnPage:currentPage});}

async function openClubLogoStudio(){
  if(currentUser()?.role!=='admin')return;
  const root=document.getElementById('modal_root'),input=document.getElementById('club_logo');if(!root||!input)return;
  const file=input.files?.[0],preview=document.getElementById('club_logo_preview');
  const source=Object.hasOwn(input,'sglLogoDraft')?input.sglLogoDraft:file?await fileToDataUrl(file):preview?.getAttribute('src')||'';
  if(document.getElementById('club_logo')!==input)return;
  const nodes=[...root.childNodes],name=value('club_name').trim()||'Logo del club';
  const restore=()=>{photoStudioTarget=null;profileStudioImage=null;root.replaceChildren(...nodes);};
  openProfilePhotoStudio({type:'club',name,source,onCancel:restore,onSave:data=>{input.sglLogoDraft=data;input.value='';restore();if(preview)preview.src=data;toast('Logo preparado. Pulsa Guardar Club para confirmar.');}});
  document.getElementById('profile_file')?.removeAttribute('capture');
  const save=document.getElementById('save_profile_photo');if(save)save.textContent='Usar este logo';
}

async function loadProfileStudioFile(file){if(!file)return;await loadProfileStudioSource(await fileToDataUrl(file));}
async function loadProfileStudioSource(source){profileStudioImage=await loadImage(source);profileStudioOffset={x:0,y:0};document.getElementById('save_profile_photo')?.removeAttribute('disabled');renderProfileStudio();}
function profileStudioGeometry(){const canvas=document.getElementById('profile_canvas');if(!canvas||!profileStudioImage)return {maxX:0,maxY:0};const rotated=Math.abs(profileStudioRotation)%180===90,sourceW=rotated?profileStudioImage.naturalHeight:profileStudioImage.naturalWidth,sourceH=rotated?profileStudioImage.naturalWidth:profileStudioImage.naturalHeight,zoom=Number(value('photo_zoom')||100)/100,scale=(photoStudioTarget?.type==='club'?Math.min:Math.max)(canvas.width/sourceW,canvas.height/sourceH)*zoom;return {maxX:Math.max(0,(sourceW*scale-canvas.width)/2),maxY:Math.max(0,(sourceH*scale-canvas.height)/2)};}
function clampProfileStudioOffset(){const {maxX,maxY}=profileStudioGeometry();profileStudioOffset.x=Math.max(-maxX,Math.min(maxX,profileStudioOffset.x));profileStudioOffset.y=Math.max(-maxY,Math.min(maxY,profileStudioOffset.y));}
function setProfileStudioPosition(x,y){const {maxX,maxY}=profileStudioGeometry();profileStudioOffset={x:Number(x)*maxX,y:Number(y)*maxY};renderProfileStudio();}
function bindProfileStudioDrag(){const canvas=document.getElementById('profile_canvas');if(!canvas)return;canvas.addEventListener('pointerdown',event=>{if(!profileStudioImage)return;const rect=canvas.getBoundingClientRect();profileStudioDrag={pointerId:event.pointerId,x:event.clientX,y:event.clientY,scaleX:canvas.width/rect.width,scaleY:canvas.height/rect.height};canvas.setPointerCapture?.(event.pointerId);canvas.classList.add('dragging');event.preventDefault();});canvas.addEventListener('pointermove',event=>{if(!profileStudioDrag||profileStudioDrag.pointerId!==event.pointerId)return;profileStudioOffset.x+=(event.clientX-profileStudioDrag.x)*profileStudioDrag.scaleX;profileStudioOffset.y+=(event.clientY-profileStudioDrag.y)*profileStudioDrag.scaleY;profileStudioDrag.x=event.clientX;profileStudioDrag.y=event.clientY;renderProfileStudio();event.preventDefault();});const finish=event=>{if(!profileStudioDrag||profileStudioDrag.pointerId!==event.pointerId)return;profileStudioDrag=null;canvas.classList.remove('dragging');};canvas.addEventListener('pointerup',finish);canvas.addEventListener('pointercancel',finish);canvas.addEventListener('lostpointercapture',()=>{profileStudioDrag=null;canvas.classList.remove('dragging');});}
function rotateProfileStudio(degrees){profileStudioRotation=(profileStudioRotation+degrees)%360;profileStudioOffset={x:0,y:0};renderProfileStudio();}
function applyPhotoPreset(name){const presets={natural:[100,100,100,0,0,0,0,0],vivid:[104,116,135,0,0,0,0,4],soft:[108,88,90,0,0,0,1,0],warm:[104,105,118,0,8,8,0,0],cool:[102,108,105,0,0,-10,0,0],portrait:[108,96,92,0,3,0,.5,8],contrast:[100,138,108,0,0,0,0,5],vintage:[96,112,75,0,35,-8,.3,35],bw:[104,118,0,100,0,0,0,8],sepia:[100,108,75,0,80,4,0,15]},preset=presets[name]||presets.natural,ids=['photo_brightness','photo_contrast','photo_saturation','photo_gray','photo_sepia','photo_hue','photo_blur','photo_vignette'];ids.forEach((id,index)=>{const root=document.getElementById(id);if(root)root.value=preset[index];});renderProfileStudio();}
function resetProfileStudio(){profileStudioRotation=0;profileStudioOffset={x:0,y:0};const values={photo_zoom:100,photo_brightness:100,photo_contrast:100,photo_saturation:100,photo_gray:0,photo_sepia:0,photo_hue:0,photo_blur:0,photo_vignette:0};for(const [id,next] of Object.entries(values)){const root=document.getElementById(id);if(root)root.value=next;}renderProfileStudio();}
function renderProfileStudio(){const canvas=document.getElementById('profile_canvas');if(!canvas)return;const outputs={photo_zoom:'%',photo_brightness:'%',photo_contrast:'%',photo_saturation:'%',photo_gray:'%',photo_sepia:'%',photo_hue:'°',photo_blur:'',photo_vignette:'%'};for(const [id,suffix] of Object.entries(outputs)){const output=document.getElementById(`${id}_value`),root=document.getElementById(id);if(output&&root)output.value=`${root.value}${suffix}`;}const ctx=canvas.getContext('2d');ctx.clearRect(0,0,canvas.width,canvas.height);if(photoStudioTarget?.type!=='club'){ctx.fillStyle='#edf0f4';ctx.fillRect(0,0,canvas.width,canvas.height);}if(!profileStudioImage){ctx.fillStyle='#667085';ctx.font='24px Arial';ctx.textAlign='center';ctx.fillText('Selecciona una fotografía',canvas.width/2,canvas.height/2);return;}const rotated=Math.abs(profileStudioRotation)%180===90,sourceW=rotated?profileStudioImage.naturalHeight:profileStudioImage.naturalWidth,sourceH=rotated?profileStudioImage.naturalWidth:profileStudioImage.naturalHeight,zoom=Number(value('photo_zoom')||100)/100,scale=(photoStudioTarget?.type==='club'?Math.min:Math.max)(canvas.width/sourceW,canvas.height/sourceH)*zoom;clampProfileStudioOffset();ctx.imageSmoothingEnabled=true;ctx.imageSmoothingQuality='high';ctx.save();ctx.translate(canvas.width/2+profileStudioOffset.x,canvas.height/2+profileStudioOffset.y);ctx.rotate(profileStudioRotation*Math.PI/180);ctx.filter=`brightness(${value('photo_brightness')||100}%) contrast(${value('photo_contrast')||100}%) saturate(${value('photo_saturation')||100}%) grayscale(${value('photo_gray')||0}%) sepia(${value('photo_sepia')||0}%) hue-rotate(${value('photo_hue')||0}deg) blur(${value('photo_blur')||0}px)`;ctx.drawImage(profileStudioImage,-profileStudioImage.naturalWidth*scale/2,-profileStudioImage.naturalHeight*scale/2,profileStudioImage.naturalWidth*scale,profileStudioImage.naturalHeight*scale);ctx.restore();const vignette=Number(value('photo_vignette')||0)/100;if(vignette){const gradient=ctx.createRadialGradient(300,300,160,300,300,430);gradient.addColorStop(0,'rgba(0,0,0,0)');gradient.addColorStop(1,`rgba(0,0,0,${Math.min(.7,vignette)})`);ctx.fillStyle=gradient;ctx.fillRect(0,0,600,600);}}
async function saveProfileStudio(){
  if(!profileStudioImage)return;renderProfileStudio();const canvas=document.getElementById('profile_canvas'),data=canvas.toDataURL(photoStudioTarget?.type==='club'?'image/png':'image/jpeg',.95),target=photoStudioTarget||{type:'user',userId:currentUser()?.id,returnPage:'profile'};
  if(target.type==='club'){target.onSave(data);return;}
  if(target.type==='member'){
    const item=accessibleClasses().find(entry=>entry.id===target.classId),member=item?classWork(item).members.find(entry=>entry.id===target.memberId):null;if(!item||!member)return toast('No se encontró el Conquistador.');member.photo=data;await persist(true,{action:'Foto de Conquistador actualizada',entity:item.id,detail:member.fullName});closeModal();photoStudioTarget=null;renderClassWorkspace();toast('Foto del Conquistador guardada.');return;
  }
  const actor=currentUser(),user=userById(target.userId);if(!user||(!actor||actor.id!==user.id)&&actor?.role!=='admin')return toast('No tienes permiso para editar esta foto.');user.photo=data;await persist(true,{action:'Foto de perfil actualizada',entity:user.id,detail:'Estudio de foto'});closeModal();photoStudioTarget=null;
  if(actor.id===user.id){renderShell();go('profile');}else if(target.returnPage==='aspirants')renderAspirants();else renderUsers();toast('Foto de perfil guardada.');
}
async function makeSquareProfilePhoto(file){const source=await fileToDataUrl(file),image=await loadImage(source),canvas=document.createElement('canvas'),ctx=canvas.getContext('2d');canvas.width=600;canvas.height=600;const scale=Math.max(600/image.naturalWidth,600/image.naturalHeight),w=image.naturalWidth*scale,h=image.naturalHeight*scale;ctx.drawImage(image,(600-w)/2,(600-h)/2,w,h);return canvas.toDataURL('image/jpeg',.95);}
async function changeOwnPassword() { const password=prompt('Nueva contraseña (mínimo 8 caracteres):',''); if(!password||password.length<8)return toast('La contraseña requiere al menos 8 caracteres.'); const user=currentUser(); user.password=await passwordRecord(password); await persist(true,{action:'Contraseña actualizada',entity:user.id,detail:''}); toast('Contraseña actualizada.'); }

function settingsBackupZone(user=currentUser()){
  const scope=user.role==='admin'?'todos los usuarios, clubes y expedientes':['aspirant','pastor_aspirant'].includes(user.role)?'tu expediente, informes y archivos':'los expedientes de tus Aspirantes asignados';
  return `<section class="card settings-backup-zone" style="margin-top:16px"><div class="card-head"><div><h3>Copias de seguridad</h3><p class="subtitle" style="margin:0">Incluye ${scope}.</p></div></div><div class="backup-options"><div class="backup-card"><h4>Copia protegida</h4><p>Archivo cifrado para PC, USB o disco externo.</p><button class="btn primary block" onclick="exportEncryptedBackup()">Crear copia</button></div><div class="backup-card"><h4>Copia JSON</h4><p>Datos estructurados para migración o revisión técnica.</p><button class="btn ghost block" onclick="exportJsonBackup()">Descargar JSON</button></div><div class="backup-card"><h4>Google Drive</h4><p>Usa la carpeta personal o administrativa habilitada.</p><button class="btn ghost block" onclick="go('drive')">Abrir Drive</button></div></div>${user.role==='admin'?`<div class="settings-restore"><b>Restaurar copia protegida</b><label class="btn ghost small">Elegir copia<input hidden type="file" accept=".cabackup" onchange="restoreEncryptedBackup(this.files[0]);this.value=''"/></label></div>`:''}</section>`;
}
function settingsConflictZone(user=currentUser()){
  if(user.role!=='admin')return '';
  const count=database.conflicts?.length||0;
  return `<section class="card settings-conflict-zone" style="margin-top:16px"><div class="card-head"><div><h3>Conflictos de sincronización</h3><p class="subtitle" style="margin:0">${count?`${count} conflicto${count===1?'':'s'} requiere${count===1?'':'n'} revisión.`:'No hay conflictos pendientes.'}</p></div><button class="btn ${count?'warn':'ghost'}" onclick="go('conflicts')">${count?'Revisar conflictos':'Ver historial'}</button></div></section>`;
}

function renderSettings() {
  const user=currentUser();setTop('Configuración',ROLE_LABELS[user.rawRole624||user.role]); const settings=database.settings;if(user.role!=='admin'){setContent(`<div class="page-head"><div><div class="eyebrow">Preferencias y respaldo</div><h1 class="page-title">Configuración</h1><p class="subtitle">Tu tema se cambia desde Mi perfil. Aquí puedes gestionar tus copias.</p></div></div>${settingsBackupZone(user)}`);return;}
  const activities=visibleActivities(),resetZone=currentUser()?.role==='admin'?`<div class="card factory-reset-zone" style="margin-top:16px"><div class="card-head"><div><h3>Administración de datos</h3><p class="subtitle" style="margin:0">Inicia una instalación limpia conservando las clases, requisitos y recursos estructurales del programa.</p></div><button class="btn red" onclick="openFactoryReset()">Empezar desde cero</button></div><div class="notice warn" style="margin-top:12px"><b>Copia obligatoria.</b> El sistema no eliminará ningún dato hasta que guardes correctamente una copia protegida completa.</div></div>`:'';setTimeout(decorateGoogleSettings,0);
  setContent(`<div class="page-head"><div><div class="eyebrow">Servidor y servicios</div><h1 class="page-title">Configuración general</h1><p class="subtitle">Los cambios locales funcionan sin servidor. La URL habilita sincronización multiusuario.</p></div></div><form class="two" onsubmit="saveSettings(event)"><div class="card"><h3>Sincronización e informes</h3><label class="field-group">URL del servidor<input id="set_server" class="field" placeholder="https://sistemagestionliderazgo.ejemplo.org" value="${esc(settings.serverUrl||'')}"></label><label class="field-group">Token de Administrador<input id="set_token" class="field" type="password" value="${esc(settings.serverToken||'')}"></label><label class="choice"><input id="set_auto" type="checkbox" ${settings.autoSync?'checked':''}> Sincronizar automáticamente al recuperar internet</label><label class="field-group">Letra predeterminada para nuevas cuentas<select id="set_font" class="field"><option ${settings.reportFont==='Arial'?'selected':''}>Arial</option><option ${settings.reportFont==='Times New Roman'?'selected':''}>Times New Roman</option></select></label><label class="field-group">Tema de color<select id="set_theme" class="field" onchange="previewTheme(this.value)">${Object.entries(THEME_LABELS).map(([id,label])=>`<option value="${id}" ${user.theme===id?'selected':''}>${label}</option>`).join('')}</select></label><div class="theme-swatches">${Object.keys(THEME_LABELS).map(id=>`<button type="button" class="theme-swatch theme-${id}" title="${esc(THEME_LABELS[id])}" onclick="document.getElementById('set_theme').value='${id}';previewTheme('${id}')"></button>`).join('')}</div><div class="notice">Cada Aspirante puede elegir Arial o Times New Roman en Mi perfil. Carta 21,6 × 27,9 cm · vista y PDF con la misma paginación · márgenes laterales 3/2 cm.</div></div><div class="card"><h3>Logo de ingreso e interfaz</h3><div class="user-row" style="margin-bottom:12px"><img class="club-logo" src="${appLogo()}"><p class="subtitle" style="margin:0">Visible para Administrador, Tutor y Aspirante. No se usa en informes.</p></div><label class="field-group">Cambiar logo PNG<input id="set_app_logo" type="file" class="field" accept="image/png,image/*"></label><button type="button" class="btn ghost small" onclick="resetAppLogo()">Restaurar logo de Música</button><hr style="border:0;border-top:1px solid var(--border);margin:18px 0"><h3>Google OAuth</h3><label class="field-group">Client ID<input id="set_google" class="field" value="${esc(settings.googleClientId||'')}"></label><div class="notice">Permiso recomendado: <b>drive.file</b>. La carpeta es personal y Google no crea usuarios.</div><button class="btn primary" style="margin-top:14px">Guardar configuración</button></div></form>${settingsBackupZone(user)}${settingsConflictZone(user)}${resetZone}<div class="card" style="margin-top:16px"><div class="card-head"><div><h3>Registro de actividad</h3><p class="subtitle" style="margin:0">Se conserva hasta que el Administrador decida borrarlo.</p></div><div class="actions"><button class="btn ${activityRange==='today'?'primary':'ghost'} small" onclick="setActivityRange('today')">Hoy</button><button class="btn ${activityRange==='all'?'primary':'ghost'} small" onclick="setActivityRange('all')">Todo</button><button class="btn red small" onclick="confirmClearActivity()">Borrar lo mostrado</button></div></div><div class="table-wrap"><table class="table"><thead><tr><th>Fecha</th><th>Persona</th><th>Acción</th><th>Detalle</th></tr></thead><tbody>${activities.slice(0,250).map(item=>`<tr><td>${formatDateTime(item.at)}</td><td>${esc(item.actorName)}</td><td>${esc(item.action)}</td><td>${esc(item.detail)}</td></tr>`).join('')||'<tr><td colspan="4" class="empty">Sin actividad en este período.</td></tr>'}</tbody></table></div></div><div class="card" style="margin-top:16px"><div class="card-head"><div><h3>Diagnóstico de cierres</h3><p class="subtitle" style="margin:0">Los errores se guardan localmente para investigar cierres inesperados.</p></div><button class="btn red small" onclick="confirmClearErrors()">Borrar diagnóstico</button></div><div class="table-wrap"><table class="table"><thead><tr><th>Fecha</th><th>Tipo</th><th>Mensaje</th></tr></thead><tbody>${database.errors.slice(0,30).map(item=>`<tr><td>${formatDateTime(item.at)}</td><td>${esc(item.kind)}</td><td>${esc(item.message)}</td></tr>`).join('')||'<tr><td colspan="3" class="empty">No se registraron errores.</td></tr>'}</tbody></table></div></div>`);
}

function decorateGoogleSettings(){ return false; }
function openGoogleSetup(){ return false; }

function leadershipRecognition(score, actions){
  const messages=[
    '🌟 ¡Excelente trabajo! Tu constancia y participación reflejan el compromiso de un gran líder.',
    '👏 ¡Felicitaciones por tu avance! Cada informe, actividad y aprendizaje fortalece tu liderazgo.',
    '🚀 ¡Sigue creciendo! Tu dedicación demuestra que estás avanzando para ser un gran Guía Mayor Máster.',
    '🔎 ¡Gran iniciativa! Explorar nuevas funciones y aprender demuestra una actitud de liderazgo.',
    '📚 ¡Muy buen progreso! Tu esfuerzo constante construye un camino de servicio y excelencia.',
    '🏅 ¡Tu compromiso inspira! Continúa desarrollando tus habilidades como líder.'
  ];
  return messages[Math.floor(Math.random()*messages.length)];
}

function classLeadershipRecognition(user){
  const hasClass=(database.classes||[]).some(c=>c.instructorId===user?.id || c.leaderId===user?.id);
  if(!hasClass)return '';
  const messages=[
    '👥 ¡Felicitaciones por tu servicio! Tu dedicación acompañando a niños y adolescentes deja una huella importante.',
    '🌱 Tu liderazgo en la clase ayuda a formar nuevas generaciones. Gracias por enseñar y guiar con compromiso.',
    '👏 Excelente trabajo como instructor. Cada clase compartida fortalece el crecimiento de tus alumnos.'
  ];
  return messages[Math.floor(Math.random()*messages.length)];
}

function leadershipScore(items,userId){
  const own=items.filter(x=>x.actorId===userId);
  const participation=own.length;
  const navigation=own.filter(x=>/abr|ver|consult|naveg|entr|open/i.test(x.action+' '+x.detail)).length;
  const functions=own.filter(x=>/crear|modificar|editar|aprobar|guardar|completar|prueba/i.test(x.action+' '+x.detail)).length;
  return participation*0.4+navigation*0.25+functions*0.2;
}

function renderLeadership(){
  const user=currentUser();
  const from=document.getElementById('lead_from')?.value || leadershipRange.from || localDate();
  const to=document.getElementById('lead_to')?.value || leadershipRange.to || localDate();
  const start=new Date(from+'T00:00:00'), end=new Date(to+'T23:59:59');
  const logs=(database.audit||[]).filter(a=>{const d=new Date(a.at);return d>=start&&d<=end;});
  const users=accessibleAspirants(user).concat(user.role==='admin'?database.users.filter(u=>u.role==='tutor'):[]);
  const ranks=users.map(u=>({u,score:Math.round(leadershipScore(logs,u.id)*10)/10,actions:logs.filter(a=>a.actorId===u.id).length})).sort((a,b)=>b.score-a.score).slice(0,10);
  const me=ranks.find(r=>r.u.id===user.id)||ranks[0];
  setTop('Liderazgo en Acción','Reconocimiento y evaluación por período');
  setContent(`<div class="page-head"><div><h1 class="page-title">🏆 Liderazgo en Acción</h1><p class="subtitle">Reconocemos participación, navegación, aprendizaje y compromiso.</p></div></div>
  <div class="card"><h3>Periodo de evaluación</h3><div class="two"><label class="field-group">Desde<input id="lead_from" class="field" type="date" value="${from}"></label><label class="field-group">Hasta<input id="lead_to" class="field" type="date" value="${to}"></label></div><button class="btn primary" onclick="leadershipApplyPeriod()">Analizar periodo</button></div>
  <div class="grid4"><div class="card kpi"><small>Acciones</small><strong>${logs.length}</strong></div><div class="card kpi"><small>Participantes</small><strong>${ranks.length}</strong></div><div class="card kpi"><small>Desde</small><strong>${from}</strong></div><div class="card kpi"><small>Hasta</small><strong>${to}</strong></div></div>
  ${me?`<div class="card"><h3>🎉 Reconocimiento personal</h3><p>${esc(leadershipRecognition(me.score,me.actions))}</p><div class="notice">Tu desempeño en este periodo: <b>${me.score} puntos</b> · ${me.actions} acciones registradas.</div>${classLeadershipRecognition(me.u)?`<div class="notice">${esc(classLeadershipRecognition(me.u))}</div>`:''}</div>`:''}
  <div class="card"><h3>📌 Mi actividad del período</h3><div class="grid4"><div class="kpi"><small>Navegaciones</small><strong>${me?logs.filter(a=>a.actorId===me.u.id && /abr|ver|consult|naveg|entr|open/i.test(a.action+' '+a.detail)).length:0}</strong></div><div class="kpi"><small>Informes y acciones</small><strong>${me?logs.filter(a=>a.actorId===me.u.id).length:0}</strong></div><div class="kpi"><small>Pruebas/completados</small><strong>${me?logs.filter(a=>a.actorId===me.u.id && /prueba|complet|evalu/i.test(a.action+' '+a.detail)).length:0}</strong></div><div class="kpi"><small>Funciones usadas</small><strong>${me?logs.filter(a=>a.actorId===me.u.id && /crear|modificar|editar|guardar|aprobar/i.test(a.action+' '+a.detail)).length:0}</strong></div></div></div>
  <div class="card"><h3>🎓 Camino Guía Mayor</h3><p>El avance de Guía Mayor se calcula según las tarjetas, requisitos, actividades y evidencias asignadas.</p><div class="notice">🏆 Guía Mayor Máster: recibirás reconocimientos por tu crecimiento, servicio y liderazgo.</div></div>
  <div class="card"><h3>🏆 Ranking de Liderazgo</h3>${ranks.map((r,i)=>`<div class="user-row"><b>#${i+1} ${esc(r.u.fullName)}</b><span>${r.score} puntos · ${r.actions} acciones</span></div>`).join('')||'<div class="empty">Sin actividad en este periodo.</div>'}</div>`);
}
function leadershipApplyPeriod(){leadershipRange={from:document.getElementById('lead_from').value,to:document.getElementById('lead_to').value};renderLeadership();}

function visibleActivities(){if(activityRange==='all')return database.audit;const today=localDate();return database.audit.filter(item=>{const date=new Date(item.at),offset=date.getTimezoneOffset();return new Date(date.getTime()-offset*60000).toISOString().slice(0,10)===today;});}
function setActivityRange(range){activityRange=range==='all'?'all':'today';renderSettings();}
function confirmClearActivity(){const count=visibleActivities().length;if(!count)return toast('No hay actividad para borrar.');const label=activityRange==='all'?'todo el registro de actividad':'la actividad de hoy';showModal(`<div class="modal small"><div class="modal-head"><div><h3>Borrar actividad</h3><p>Esta acción no elimina usuarios, informes ni avances.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="notice warn">Se borrará ${esc(label)}: <b>${count}</b> registro${count===1?'':'s'}.</div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn red" onclick="clearVisibleActivity()">Borrar definitivamente</button></div></div>`);}
async function clearVisibleActivity(){const shown=new Set(visibleActivities().map(item=>item.id));if(!shown.size)return toast('No hay actividad para borrar.');database.audit=database.audit.filter(item=>!shown.has(item.id));await persist(true);await writeNativeSnapshot();closeModal();renderSettings();toast(`${shown.size} registro${shown.size===1?'':'s'} de actividad eliminado${shown.size===1?'':'s'}.`);}
function confirmClearErrors(){if(!database.errors.length)return toast('No hay diagnóstico para borrar.');showModal(`<div class="modal small"><div class="modal-head"><div><h3>Borrar diagnóstico</h3><p>No se eliminarán datos de usuarios.</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="notice warn">Se borrarán <b>${database.errors.length}</b> mensajes técnicos.</div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn red" onclick="clearErrors()">Borrar definitivamente</button></div></div>`);}
async function clearErrors(){const count=database.errors.length;database.errors=[];await persist(true);await writeNativeSnapshot();closeModal();renderSettings();toast(`${count} mensaje${count===1?'':'s'} de diagnóstico eliminado${count===1?'':'s'}.`);}

async function createMandatoryResetBackup(password){
  if(currentUser()?.role!=='admin')return {ok:false,error:'Solo el Administrador puede reiniciar la instalación.'};
  if(!password||password.length<8)return {ok:false,error:'La contraseña de la copia requiere al menos 8 caracteres.'};
  try{const salt=crypto.getRandomValues(new Uint8Array(16)),iv=crypto.getRandomValues(new Uint8Array(12)),raw=await deriveKeyBytes(password,salt,180000,32),key=await crypto.subtle.importKey('raw',raw,{name:'AES-GCM'},false,['encrypt']),payload=await backupPayload(),encrypted=await crypto.subtle.encrypt({name:'AES-GCM',iv},key,new TextEncoder().encode(JSON.stringify(payload))),envelope={format:'MinisterioMusicaEncryptedBackup',version:1,kdf:'PBKDF2-SHA256',iterations:180000,salt:toBase64(salt),iv:toBase64(iv),ciphertext:toBase64(new Uint8Array(encrypted))},stamp=new Date().toISOString().replace(/[:.]/g,'-'),result=await saveBlobWithDialog(`MinisterioMusica_ANTES_DE_REINICIAR_${stamp}.cabackup`,new Blob([JSON.stringify(envelope)],{type:'application/octet-stream'}),'Copia protegida|*.cabackup');if(!result?.ok)return {ok:false,canceled:!!result?.canceled,error:result?.canceled?'Guardado cancelado. No se modificó ningún dato.':'La copia no pudo guardarse. No se modificó ningún dato.'};return result;}catch(error){await logClientError('Copia previa al reinicio',error);return{ok:false,error:`No se pudo crear la copia: ${error.message}`};}
}
function openFactoryReset(){
  if(currentUser()?.role!=='admin')return;
  showModal(`<div class="modal small"><div class="modal-head"><div><h3>Empezar desde cero</h3><p>Paso 1 de 2 · Guardar copia completa antes de borrar</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="notice warn"><b>No se borrará nada todavía.</b> Primero guarda una copia protegida de los informes, tareas, archivos y registros de Música.</div><label class="field-group" style="margin-top:14px">Contraseña para proteger la copia<input id="factory_reset_password" class="field" type="password" minlength="8" autocomplete="new-password" placeholder="Mínimo 8 caracteres"></label><div id="factory_reset_error" class="error"></div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cancelar</button><button id="factory_reset_backup_btn" class="btn red" onclick="confirmFactoryResetBackup()">Guardar copia y continuar</button></div></div>`);
}
async function confirmFactoryResetBackup(){
  const password=value('factory_reset_password'),errorRoot=document.getElementById('factory_reset_error'),button=document.getElementById('factory_reset_backup_btn');if(password.length<8){if(errorRoot)errorRoot.textContent='Usa al menos 8 caracteres para proteger la copia.';return;}if(button){button.disabled=true;button.textContent='Guardando copia…';}const result=await createMandatoryResetBackup(password);if(!result?.ok){if(errorRoot)errorRoot.textContent=result.error||'No se guardó la copia. El reinicio fue cancelado.';if(button){button.disabled=false;button.textContent='Guardar copia y continuar';}return;}showModal(`<div class="modal small"><div class="modal-head"><div><h3>Confirmar instalación nueva</h3><p>Paso 2 de 2 · La copia ya fue guardada</p></div><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><div class="notice"><b>Copia creada correctamente.</b><br>${esc(result.path||'Archivo guardado')}</div><p class="subtitle" style="margin:16px 0 8px">Se reiniciarán los expedientes, tareas, materiales e indicaciones de Música. Se conservarán las cuentas, eventos, certificados y la conexión con SGL.</p><label class="field-group">Escribe <b>NUEVO</b> para confirmar<input id="factory_reset_confirm" class="field" autocomplete="off" placeholder="NUEVO"></label><div id="factory_reset_final_error" class="error"></div></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Conservar datos</button><button class="btn red" onclick="executeFactoryReset()">Reiniciar expedientes de Música</button></div></div>`);
}
async function executeFactoryReset(confirmation=''){
  if(currentUser()?.role!=='admin')return;const typed=(confirmation||value('factory_reset_confirm')).trim().toUpperCase();if(typed!=='NUEVO'){setText('factory_reset_final_error','Escribe NUEVO exactamente para confirmar.');return;}
  const safety={format:'MinisterioMusicaResetSafety',version:1,savedAt:isoNow(),database:structuredClone(database),files:await allFiles()};await kvSet(`safety_${Date.now()}`,safety);await clearFiles();database=blankDatabase();await kvSet('database',database);await writeNativeSnapshot();sessionStorage.removeItem('musica_session_user');localStorage.removeItem('musica_session_user');localStorage.removeItem('musica_emergency_draft');activeAspirantId=null;activeClassId='';activeClassSection='dashboard';classMemberSearch='';classMemberFilter='active';classTrackFilter='all';classPlanFilter='all';reportPdfCache.clear();applyTheme();renderBootstrap();
}

function previewTheme(theme){document.documentElement.dataset.theme=THEME_LABELS[theme]?theme:'classic_blue';}
async function saveSettings(event) { event.preventDefault(); database.settings.serverUrl=value('set_server').trim().replace(/\/$/,''); database.settings.serverToken=value('set_token').trim(); database.settings.googleClientId=value('set_google').trim(); database.settings.reportFont=value('set_font')||'Arial'; const actor=currentUser();if(actor)actor.theme=THEME_LABELS[value('set_theme')]?value('set_theme'):'classic_blue'; database.settings.theme='classic_blue'; database.settings.autoSync=document.getElementById('set_auto').checked; const logo=document.getElementById('set_app_logo')?.files?.[0];if(logo)database.settings.musicAppLogo=await imageToTransparentDataUrl(logo,1200);await persist(true,{action:'Configuración actualizada',entity:'settings',detail:database.settings.serverUrl||'Solo local'}); renderShell(); toast('Configuración guardada.'); }
async function resetAppLogo(){database.settings.musicAppLogo='';await persist(true,{action:'Logo de ingreso restaurado',entity:'settings',detail:'Sistema de Gestión de Música'});renderShell();toast('Logo predeterminado restaurado.');}

function renderConflicts() {
  setTop('Conflictos de sincronización','Administración');
  setContent(`<div class="page-head"><div><div class="eyebrow">Protección de cambios</div><h1 class="page-title">${database.conflicts.length} conflicto${database.conflicts.length===1?'':'s'}</h1><p class="subtitle">Nunca se sobrescriben silenciosamente cambios hechos por dos personas.</p></div></div><div class="card">${database.conflicts.map(item=>`<div class="conflict"><b>${esc(item.entity||'Cambio simultáneo')}</b><p>${esc(item.message||'El servidor y este dispositivo tienen versiones distintas.')}</p><div class="actions"><button class="btn ghost small" onclick="resolveConflict('${item.id}','server')">Conservar servidor</button><button class="btn warn small" onclick="resolveConflict('${item.id}','local')">Conservar este dispositivo</button></div></div>`).join('')||'<div class="empty">No hay conflictos pendientes.</div>'}</div>`);
}

async function resolveConflict(id,choice) { const conflict=database.conflicts.find(item=>item.id===id); if(!conflict)return; if(choice==='server'&&conflict.serverDatabase)database=conflict.serverDatabase; database.conflicts=database.conflicts.filter(item=>item.id!==id); await persist(true,{action:'Conflicto resuelto',entity:id,detail:choice}); renderConflicts(); }

async function syncNow() {
  if (!database?.settings?.serverUrl) { database.sync={state:'local',lastSyncedAt:database.sync?.lastSyncedAt||null,message:'Guardado en este dispositivo'}; renderTopSync(); return; }
  if (!navigator.onLine) { database.sync={state:'pending',lastSyncedAt:database.sync?.lastSyncedAt||null,message:'Sin conexión · cambios guardados'}; renderTopSync(); return; }
  database.sync={state:'syncing',lastSyncedAt:database.sync?.lastSyncedAt||null,message:'Sincronizando…'}; renderTopSync();
  try {
    const localGeneration=Math.max(1,Number(database?.restoreGeneration||1)||1);
    const response=await fetch(`${database.settings.serverUrl}/api/sync`,{
      method:'POST',headers:{'Content-Type':'application/json','Authorization':`Bearer ${database.settings.serverToken||''}`},
      body:JSON.stringify({installationId:database.installationId,baseVersion:database.serverVersion||0,restoreGeneration:localGeneration,database})
    });
    if(response.status===409){
      const remote=await response.json().catch(()=>({}));
      if(remote?.code==='RESTORE_REQUIRED'&&typeof window.sgl5028HandleRestoreRequired==='function'){
        try{await window.sgl5028HandleRestoreRequired(remote);}catch(error){error.payload=remote;error.restoreHandled=true;throw error;}
        return;
      }
      database.conflicts.push({id:newId('conflict'),at:isoNow(),entity:'Base de datos',message:'Hay cambios remotos y locales desde la última sincronización.',serverDatabase:remote.database||null});
      database.sync={state:'conflict',lastSyncedAt:database.sync.lastSyncedAt,message:'Requiere resolver conflicto'};
      await kvSet('database',database);renderTopSync();
      if(currentUser()?.role==='admin'){if(currentPage==='dashboard')renderDashboard();else toast('Se detectó un conflicto. Revísalo desde Inicio o Configuración.');}
      return;
    }
    if(!response.ok)throw new Error(`HTTP ${response.status}`);
    const result=await response.json();
    if(result.database){
      const priorGeneration=Math.max(1,Number(database?.restoreGeneration||1)||1);
      const priorRestoreId=String(database?.restoreId||'');
      const priorRestoredAt=String(database?.restoredAt||'');
      const priorServerVersion=Number(database?.serverVersion||0)||0;
      database=result.database;
      database.restoreGeneration=Math.max(priorGeneration,Number(result.restoreGeneration||database.restoreGeneration||1)||1);
      database.restoreId=String(result.restoreId||priorRestoreId||database.restoreId||'');
      database.restoredAt=String(result.restoredAt||priorRestoredAt||database.restoredAt||'');
      database.restorePending=false;
      database.serverVersion=Math.max(priorServerVersion,Number(result.serverVersion||database.serverVersion||0)||0);
    } else if(Number(result.serverVersion||0)>0) database.serverVersion=Math.max(Number(database.serverVersion||0)||0,Number(result.serverVersion)||0);
    database.sync={state:'ok',lastSyncedAt:isoNow(),message:'Sincronizado'};
    await kvSet('database',database); renderTopSync(); if(document.getElementById('content'))render();
  } catch(error){
    if(error?.restoreHandled||error?.payload?.code==='RESTORE_REQUIRED')return;
    database.sync={state:'pending',lastSyncedAt:database.sync.lastSyncedAt,message:'Pendiente · servidor no disponible'};
    await kvSet('database',database);renderTopSync();
  }
}

function renderTopSync() {
  const root=document.getElementById('sync_pill'); if(!root||!database)return; const state=database.sync?.state||'local',className=state==='syncing'||state==='pending'?'pending':state==='ok'?'ok':state==='conflict'?'conflict':'local'; root.className=`sync-pill sync-${className}`; root.textContent=database.sync?.message||'Guardado local'; root.title=database.sync?.lastSyncedAt?`Última sincronización: ${formatDateTime(database.sync.lastSyncedAt)}`:''; sglApplyOfflineMode(sglShouldShowOfflineMode());
}

function enableSpanishProofing(root){if(!root)return;for(const field of root.querySelectorAll('textarea,input[type="text"],[contenteditable="true"]')){field.setAttribute('lang','es');field.setAttribute('spellcheck','true');field.setAttribute('autocorrect','on');}}
function setContent(html) { const root=document.getElementById('content'); if(root){root.innerHTML=html;enableSpanishProofing(root);} }
function setText(id,textValue) { const root=document.getElementById(id); if(root)root.textContent=textValue; }
function value(id) { return document.getElementById(id)?.value||''; }
function showModal(html) { const root=document.getElementById('modal_root');if(root){const alreadyOpen=!!root.querySelector('.modal-bg');if(!alreadyOpen)sglPushTransientHistory('modal');root.innerHTML=`<div class="modal-bg" role="dialog" aria-modal="true">${html}</div>`;enableSpanishProofing(root);} }
function closeModal(fromHistory=false) { if(photoStudioTarget?.type==='club'){const target=photoStudioTarget;photoStudioTarget=null;if(!fromHistory){target.onCancel();return;}}saveEmergencyDraft(); const root=document.getElementById('modal_root'),wasOpen=!!root?.querySelector('.modal-bg'); if(wasOpen&&!fromHistory&&history.state?.sgl&&history.state.kind==='modal')sglDiscardTransientHistory(); if(root)root.innerHTML='';if(materialPreviewUrl){URL.revokeObjectURL(materialPreviewUrl);materialPreviewUrl='';} }
function toast(message) { const root=document.getElementById('toast_root'); if(!root)return; const id=newId('toast'); root.insertAdjacentHTML('beforeend',`<div id="${id}" class="toast">${esc(message)}</div>`); setTimeout(()=>document.getElementById(id)?.remove(),4200); }
function formatDate(valueText) { if(!valueText)return 'Sin fecha'; const [y,m,d]=String(valueText).slice(0,10).split('-'); return y&&m&&d?`${d}/${m}/${y}`:valueText; }
function isBirthdayToday(user){if(!user?.birthDate)return false;const [,month,day]=user.birthDate.split('-'),now=new Date();return Number(month)===now.getMonth()+1&&Number(day)===now.getDate();}
function formatDateTime(valueText) { if(!valueText)return '—'; try{return new Date(valueText).toLocaleString('es',{dateStyle:'short',timeStyle:'short'});}catch{return valueText;} }
function safeName(textValue) { return String(textValue||'archivo').normalize('NFD').replace(/[\u0300-\u036f]/g,'').replace(/[^a-zA-Z0-9_-]+/g,'_').replace(/^_+|_+$/g,''); }
function compactRequirementText(requirement){let text=String(requirement?.official_text||requirement?.requirement||requirement?.id||'Informe').replace(/^\s*\d+[.)-]?\s*/,'').replace(/\s+del Club de Conquistadores\b/ig,'').split(/[.;:]/)[0].trim();if(text.length>58)text=text.slice(0,58).replace(/\s+\S*$/,'').trim();return text||'Informe';}
function reportBaseName(requirement){const code=requirement?.areaCode||String(requirement?.areaTitle||'').match(/^([IVX]+)/)?.[1]||'I';return `${AREA_SHORT[code]||code} - ${compactRequirementText(requirement)}`;}
function shortAspirantFileName(fullName){const parts=String(fullName||'').trim().split(/\s+/).filter(Boolean);if(parts.length<=2)return parts.join(' ');const surnameIndex=parts.length>=4?parts.length-2:1;return `${parts[0]} ${parts[surnameIndex]}`.trim();}
function reportFileName(data,extension='pdf'){const requirement=data?.requirement||{},code=String(requirement.id||requirement.areaCode||'REQ').trim(),title=compactRequirementText(requirement),person=shortAspirantFileName(data?.aspirant?.fullName||'');return `${`${code} - ${title} - ${person}`.replace(/[<>:"/\\|?*]/g,'-').replace(/\s+/g,' ').trim()}.${extension}`;}
function downloadText(name,content,type='text/plain') { const blob=new Blob([content],{type}),url=URL.createObjectURL(blob),link=document.createElement('a');link.href=url;link.download=name;document.body.appendChild(link);link.click();link.remove();setTimeout(()=>URL.revokeObjectURL(url),1000); }
function downloadBlob(name,blob){const url=URL.createObjectURL(blob),link=document.createElement('a');link.href=url;link.download=name;document.body.appendChild(link);link.click();link.remove();setTimeout(()=>URL.revokeObjectURL(url),5000);}
function fileToDataUrl(file) { return new Promise((resolve,reject)=>{const reader=new FileReader();reader.onload=()=>resolve(reader.result);reader.onerror=()=>reject(reader.error);reader.readAsDataURL(file);}); }

async function blobToBase64(blob){const bytes=new Uint8Array(await blob.arrayBuffer());let binary='';for(let offset=0;offset<bytes.length;offset+=0x8000)binary+=String.fromCharCode(...bytes.subarray(offset,offset+0x8000));return btoa(binary);}
async function nativeRequest(path,payload,timeoutMs=12000){const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),timeoutMs);try{const response=await fetch(`${NATIVE_BRIDGE}${path}`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload),signal:controller.signal});if(!response.ok)throw new Error(`Servicio local ${response.status}`);return response.json();}finally{clearTimeout(timer);}}
async function waitForNativeBridge(){for(let attempt=0;attempt<5;attempt+=1){try{const result=await nativeRequest('/health',{},1800);if(result?.ok)return true;}catch{}await new Promise(resolve=>setTimeout(resolve,250));}return false;}
async function readNativeSnapshot(){try{const result=await nativeRequest('/data/read',{});return result?.ok?result.snapshot:null;}catch{return null;}}
function scheduleNativeSnapshot(){if(REMOTE_WEB)return;clearTimeout(snapshotTimer);snapshotTimer=setTimeout(writeNativeSnapshot,5000);}
async function writeNativeSnapshot(){
  if(REMOTE_WEB)return {ok:true,skipped:true};
  if(!database)return {ok:false};
  if(snapshotWriting){snapshotAgain=true;return {ok:false,pending:true};}
  snapshotWriting=true;
  try{return await nativeRequest('/data/snapshot',{format:'MinisterioMusicaLocalSnapshot',version:1,savedAt:isoNow(),database:structuredClone(database),files:await allFiles()},60000);}
  catch{return {ok:false};}
  finally{snapshotWriting=false;if(snapshotAgain){snapshotAgain=false;scheduleNativeSnapshot();}}
}
async function saveBlobWithDialog(name,blob,filter='Todos los archivos|*.*'){
  try{return await nativeRequest('/save',{filename:name,filter,dataBase64:await blobToBase64(blob)},300000);}catch(error){
    if(window.showSaveFilePicker)try{const handle=await window.showSaveFilePicker({suggestedName:name}),writable=await handle.createWritable();await writable.write(blob);await writable.close();return {ok:true,path:handle.name};}catch(pickerError){if(pickerError?.name==='AbortError')return {ok:false,canceled:true};}
    downloadBlob(name,blob);return {ok:true,path:'la carpeta de Descargas',fallback:true};
  }
}
async function nativePrintBlob(blob,name){try{return await nativeRequest('/print',{filename:name,dataBase64:await blobToBase64(blob)},300000);}catch(error){return {ok:false,error:String(error?.message||error)};}}

async function imageToDataUrl(file,maxBytes=200*1024,maxDimension=1800) {
  return (await MusicMedia612.optimize(file,'profile',{maxBytes,maxDimension})).dataUrl;
}
async function normalizeSavedLogos(){const targets=[[database.settings,'appLogo'],...(database.clubs||[]).map(club=>[club,'logo'])];for(const [owner,key] of targets){const source=owner?.[key];if(/^data:image\//.test(source||''))try{owner[key]=await transparentLogoSource(source);}catch{}}}
async function imageToTransparentDataUrl(file,maxDimension=1200){return transparentLogoSource(await fileToDataUrl(file),maxDimension);}
async function transparentLogoSource(source,maxDimension=1200){const image=await loadImage(source),scale=Math.min(1,maxDimension/Math.max(image.naturalWidth,image.naturalHeight)),canvas=document.createElement('canvas'),ctx=canvas.getContext('2d',{willReadFrequently:true}),width=Math.max(1,Math.round(image.naturalWidth*scale)),height=Math.max(1,Math.round(image.naturalHeight*scale));canvas.width=width;canvas.height=height;ctx.clearRect(0,0,width,height);ctx.drawImage(image,0,0,width,height);const imageData=ctx.getImageData(0,0,width,height),pixels=imageData.data,seen=new Uint8Array(width*height),queue=[],white=index=>pixels[index*4+3]>8&&Math.min(pixels[index*4],pixels[index*4+1],pixels[index*4+2])>225&&Math.max(pixels[index*4],pixels[index*4+1],pixels[index*4+2])-Math.min(pixels[index*4],pixels[index*4+1],pixels[index*4+2])<28,push=index=>{if(!seen[index]&&white(index)){seen[index]=1;queue.push(index);}};for(let x=0;x<width;x++){push(x);push((height-1)*width+x);}for(let y=0;y<height;y++){push(y*width);push(y*width+width-1);}for(let cursor=0;cursor<queue.length;cursor++){const index=queue[cursor],x=index%width,y=(index/width)|0;pixels[index*4+3]=0;if(x)push(index-1);if(x+1<width)push(index+1);if(y)push(index-width);if(y+1<height)push(index+width);}ctx.putImageData(imageData,0,0);let left=width,top=height,right=-1,bottom=-1;for(let y=0;y<height;y++)for(let x=0;x<width;x++)if(pixels[(y*width+x)*4+3]>8){left=Math.min(left,x);top=Math.min(top,y);right=Math.max(right,x);bottom=Math.max(bottom,y);}if(right<left)return canvas.toDataURL('image/png');const padding=Math.round(Math.max(width,height)*.01),sx=Math.max(0,left-padding),sy=Math.max(0,top-padding),sw=Math.min(width-sx,right-left+1+padding*2),sh=Math.min(height-sy,bottom-top+1+padding*2),cropped=document.createElement('canvas');cropped.width=sw;cropped.height=sh;cropped.getContext('2d').drawImage(canvas,sx,sy,sw,sh,0,0,sw,sh);return cropped.toDataURL('image/png');}
function dataUrlBytes(data){const encoded=String(data).slice(String(data).indexOf(',')+1).replace(/\s/g,'');return Math.max(0,Math.floor(encoded.length*3/4)-(encoded.endsWith('==')?2:encoded.endsWith('=')?1:0));}

window.startMinisterioMusica=()=>init().catch(error=>{document.body.innerHTML=`<div class="screen" style="display:grid;place-items:center;padding:30px"><div class="card"><h2>No se pudo iniciar el Sistema de Gestión de Música</h2><p>${esc(error.message)}</p><p class="muted">Prueba abrirla desde el ejecutable incluido en el paquete.</p></div></div>`;});


// OFFLINE_BOOT_COMPLETO_AUTOSTART 5.0.43
// Prepara la carcasa offline real después de cargar la aplicación.
window.addEventListener('load',()=>{setTimeout(()=>{try{window.sglPrepareOfflineShell?.();}catch(e){}},1500);});

// SGL 5.47: one independent tutor assignment per leadership program.
function sgl545TutorId(person, programId='class_grouped') {
 if(person && Object.prototype.hasOwnProperty.call(person,'tutorByProgram'))return String(person.tutorByProgram?.[programId]||'');
 return String(person?.tutorId||'');
}
function sgl545HasTutor(person, tutorId, programId='') {
 // Tutor linkage is independent of the account's management role.
 // Personal enrollment and review permissions are validated separately.
 if(!person||!tutorId)return false;
 if(programId)return sgl545TutorId(person,programId)===String(tutorId);
 if(Object.prototype.hasOwnProperty.call(person,'tutorByProgram'))return Object.values(person.tutorByProgram||{}).some(id=>String(id)===String(tutorId));
 return String(person.tutorId||'')===String(tutorId);
}
