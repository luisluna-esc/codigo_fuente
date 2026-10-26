/* Local fonts are enumerated without a family filter. Only used faces travel
   to the server; their original programs are retained for printable PDFs. */
(()=>{
'use strict';
const builtin=['DejaVu Sans','DejaVu Serif','DejaVu Sans Mono'];
const catalog=[],byAsset=new Map(),pendingAssets=new Map();let sequence=0,notice='',initializing=null,initialized=false,localAttempted=false,windowsHost=false;
const actual=()=>window.SGL555?.actual?.()||currentUser();
const headers=()=>({'Authorization':'Bearer '+database.settings.serverToken,'X-SGL-Personal':window.SGL555?.personalMode?.()?'1':'0'});
const endpoint='/api/music/v610';
const clean=s=>String(s||'').replace(/["\\<>;\r\n]/g,'').trim().slice(0,100);
const flags=style=>({bold:/bold|black|heavy|semibold|demi/i.test(style),italic:/italic|oblique/i.test(style)});
function status(text){notice=text;const node=document.getElementById('music613_font_status');if(node)node.textContent=text;}
function groups(){const included=[...new Set([...builtin,...catalog.filter(f=>f.source==='bundled').map(f=>f.family)])];const installed=[...new Set(catalog.filter(f=>f.source!=='bundled').map(f=>f.family))].filter(f=>!included.includes(f)).sort((a,b)=>a.localeCompare(b,'es'));return [{label:'Incluidas en Música',families:included},{label:'Fuentes disponibles',families:installed}].filter(g=>g.families.length)}
function families(){return groups().flatMap(g=>g.families);}
function best(f){let matches=catalog.filter(r=>r.family===f.font);if(!matches.length&&/^(Arial|Georgia|Times New Roman|Verdana|Trebuchet MS|Courier New|Tahoma|Palatino Linotype)$/.test(f.font)){const family=/Georgia|Times|Palatino/.test(f.font)?'DejaVu Serif':/Courier/.test(f.font)?'DejaVu Sans Mono':'DejaVu Sans';matches=catalog.filter(r=>r.family===family)}return matches.sort((a,b)=>score(a,f)-score(b,f))[0];}
function score(record,f){return ((record.bold!==!!f.bold?2:0)+(record.italic!==!!f.italic?1:0))*10+(record.source==='bundled'?0:record.source==='server'?1:2);}
function css(f){const r=byAsset.get(f.fontAsset613)||best(f);return '"'+(r?.loaded?r.alias:clean(f.font||'Arial'))+'"';}
async function load(record){
 if(record.loaded)return record;
 if(record.loading)return record.loading;
 record.loading=(async()=>{const blob=await record.blob();if(blob.size>4*1024*1024)throw Error('La fuente '+record.family+' supera 4 MB. Elige otra variante.');record.buffer=await blob.arrayBuffer();const face=new FontFace(record.alias,record.buffer,{style:record.italic?'italic':'normal',weight:record.bold?'700':'400'});await face.load();document.fonts.add(face);record.loaded=true;return record})();
 try{return await record.loading}finally{record.loading=null}
}
function add(record){const duplicate=catalog.find(f=>f.key===record.key);if(duplicate)return duplicate;record.alias='MusicFont613_'+(++sequence);catalog.push(record);return record;}
async function local(){
 if(!window.queryLocalFonts)return false;
 try{
  // This call must be made directly from the user's click (secure context).
  const records=await window.queryLocalFonts();
  for(const f of records){const style=flags(f.style+' '+f.fullName);add({family:clean(f.family),key:f.postscriptName,source:'local',...style,blob:()=>f.blob()})}
  status(`${families().length} familias disponibles.`);return true;
 }catch(e){status('Las fuentes incluidas y las del servidor siguen disponibles.');return false}
}
async function init(){
 if(initialized)return;
 if(initializing)return initializing;
 initializing=(async()=>{const response=await fetch(endpoint+'?'+new URLSearchParams({actorId:actual().id,resource:'font-catalog'}),{headers:headers(),cache:'no-store'});const data=await response.json();if(!response.ok)throw Error(data.error||'No se pudo cargar el catálogo de fuentes.');
 windowsHost=data.hostOS==='windows'&&['localhost','127.0.0.1','[::1]'].includes(location.hostname);
 for(const record of data.fonts||[])add({...record,id:record.assetId621||null,key:record.id,blob:async()=>{const r=await fetch(endpoint+'?'+new URLSearchParams({actorId:actual().id,resource:'font-source',id:record.id}),{headers:headers(),cache:'default'});if(!r.ok){const p=await r.json().catch(()=>({}));throw Error(p.error||'No se pudo cargar la fuente.')}return r.blob()}});
 for(const record of catalog)if(record.id)byAsset.set(record.id,record);
 initialized=true;status(`${groups()[0]?.families.length||builtin.length} familias incluidas · ${families().length} disponibles.`);
 try{if(!windowsHost&&window.queryLocalFonts&&(await navigator.permissions.query({name:'local-fonts'})).state==='granted'){localAttempted=true;await local()}}catch{}
 })();try{return await initializing}finally{initializing=null}
}
async function connect(){if(!initialized||windowsHost||localAttempted||!window.queryLocalFonts)return false;localAttempted=true;return local()}
function fontInfo(buffer,fallback){
 const v=new DataView(buffer);if(v.byteLength<12)throw Error('Archivo de fuente no válido.');
 const count=v.getUint16(4);let name=null,style={bold:false,italic:false};
 for(let i=0;i<count;i++){const at=12+i*16;if(at+16>v.byteLength)throw Error('Archivo de fuente incompleto.');const tag=String.fromCharCode(...new Uint8Array(buffer,at,4)),offset=v.getUint32(at+8),length=v.getUint32(at+12);if(offset+length>v.byteLength)throw Error('Archivo de fuente incompleto.');if(tag==='name')name={offset,length};if(tag==='OS/2'&&length>=64){const bits=v.getUint16(offset+62);style={bold:!!(bits&32),italic:!!(bits&1)}}}
 let family='',subfamily='',full='';
 if(name&&name.length>=6){const n=v.getUint16(name.offset+2),strings=name.offset+v.getUint16(name.offset+4);for(let i=0;i<n;i++){const at=name.offset+6+i*12;if(at+12>name.offset+name.length)break;const platform=v.getUint16(at),id=v.getUint16(at+6),len=v.getUint16(at+8),start=strings+v.getUint16(at+10);if(start+len>name.offset+name.length||![1,2,4,16,17].includes(id))continue;const data=new Uint8Array(buffer,start,len);let value='';if(platform===0||platform===3){for(let j=0;j+1<data.length;j+=2)value+=String.fromCharCode(data[j]*256+data[j+1])}else value=new TextDecoder('latin1').decode(data);if(id===16||id===1&&!family)family=value;if(id===17||id===2&&!subfamily)subfamily=value;if(id===4)full=value}}
 const found=flags(subfamily+' '+full);return {family:clean(family||fallback.replace(/\.(ttf|otf)$/i,'')),bold:style.bold||found.bold,italic:style.italic||found.italic};
}
async function importFiles(input){
 let count=0;const errors=[];
 for(const file of input.files||[]){if(!/\.(ttf|otf)$/i.test(file.name))continue;try{if(file.size>4*1024*1024)throw Error('supera 4 MB');const buffer=await file.arrayBuffer(),info=fontInfo(buffer,file.name);add({...info,key:'import:'+file.name+':'+file.size+':'+file.lastModified,blob:async()=>file});count++}catch(e){errors.push(file.name+': '+e.message)}}
 status(`${count} archivos cargados. ${new Set(catalog.map(f=>f.family)).size} familias disponibles.`+(errors.length?' No se pudieron cargar: '+errors.slice(0,3).join('; '):''));input.value='';return count;
}
async function saved(f){
 if(!f.fontAsset613)return null;
 const id=f.fontAsset613;if(byAsset.has(id))return load(byAsset.get(id));if(pendingAssets.has(id))return pendingAssets.get(id);
 const task=(async()=>{const response=await fetch(endpoint+'?'+new URLSearchParams({actorId:actual().id,resource:'font',id}),{headers:headers(),cache:'default'});
 if(!response.ok)throw Error('No se pudo abrir la fuente '+f.font+'. Restaura el respaldo completo o vuelve a cargarla.');
 const blob=await response.blob(),info=fontInfo(await blob.arrayBuffer(),f.font),record={...info,family:clean(f.font||info.family),key:'saved:'+id,alias:'MusicFont613_'+(++sequence),blob:async()=>blob,id};
 byAsset.set(id,record);catalog.push(record);await load(record);return record})();pendingAssets.set(id,task);
 try{return await task}finally{pendingAssets.delete(id)}
}
async function prepare(f){if(f.fontAsset613)return saved(f);const record=best(f);return record?load(record):null;}
async function ensure(draft,{preview=false}={}){
 await init();
 const fields=[...draft.fields,...Object.values(draft.signatures613||{})].flatMap(f=>[f,...(f.runs618||[])]),prepared=[];
 await Promise.all(fields.map(async f=>{if(f.fontAsset613){if(!preview)await saved(f);return}const record=best(f);if(record){if(!preview||!record.id)await load(record);prepared.push([f,record])}}));
 const unique=[...new Set(prepared.map(([,record])=>record))];
 if(unique.reduce((sum,r)=>sum+(r.buffer?.byteLength||0),0)>18*1024*1024)throw Error('Las fuentes usadas superan 18 MB. Utiliza menos variantes.');
 if(new Set([...fields.map(f=>f.fontAsset613).filter(Boolean),...unique.map(r=>r.id||r.key)]).size>12)throw Error('Usa hasta 12 variantes de fuente por plantilla.');
 for(const record of unique){
  if(!record.id){const dataUrl=await fileToDataUrl(new Blob([record.buffer],{type:'font/ttf'}));const {payload}=await v3Api(endpoint,{method:'POST',headers:headers(),timeout:120000,body:{actorId:actual().id,action:'font.save',value:{dataUrl},restoreGeneration:database.restoreGeneration||1}});if(!payload.ok)throw Error(payload.error||'No se pudo guardar la fuente.');record.id=payload.id;byAsset.set(record.id,record)}
 }
 for(const [f,record]of prepared)f.fontAsset613=record.id;
}
function changed(f){delete f.fontAsset613;}
window.MusicFonts613={families,groups,init,connect,css,local,importFiles,prepare,ensure,changed,status,notice:()=>notice};
})();
