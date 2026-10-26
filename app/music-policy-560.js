/* AJUSTE FUNCIONES.xlsx, rows 3–22. Music permissions shared by both entry points. */
'use strict';
(function(root){
 const keys=['review','recommend','final','notices','personal','createTutor','tutors','instructions','profile','registry','events','annual','templates','investOthers'];
 const rows={
  admin:               [1,1,1,1,1,1,1,1,1,1,1,1,1,1],
  departamental:       [1,0,1,1,1,1,1,1,1,1,1,1,1,1],
  secretaria:          [0,1,0,1,1,1,1,1,1,1,1,1,1,1],
  coordinador_musica:  [1,1,0,1,1,1,1,1,1,1,1,0,1,1],
  administrativo:     [1,1,0,1,1,1,1,0,0,0,1,0,0,0],
  regional:           [1,1,0,1,1,1,1,0,0,0,0,0,0,0],
  tutor:              [1,0,0,1,1,0,0,0,0,0,0,0,0,0],
  pastor_aspirant:     [],
  aspirant:            []
 };
 const policies=Object.fromEntries(Object.entries(rows).map(([role,row])=>[role,Object.fromEntries(keys.map((k,i)=>[k,!!row[i]]))]));
 const role=u=>{const r=typeof u==='string'?u:u?.role;return ({coordinador_sgm:'admin',administrativo_sgm:'secretaria',lider_musica:'tutor'})[r]||r;};
 const cap=(u,key)=>!!policies[role(u)]?.[key];
 function region(v){const s=String(v??'').trim().toUpperCase().replace(/^REGI[OÓ]N\s*/,'');if(s==='MBOS')return s;const n=Number(s);return /^\d+$/.test(s)&&n>=1&&n<=12?String(n):'';}
 const sameRegion=(a,b)=>!!region(a?.region)&&region(a.region)===region(b?.region);
 const active=u=>!!u&&!root.SGL571?.isDisabled?.(u.id)&&u.active!==false&&!['inactive','suspended','deleted'].includes(u.accountState)&&!u.musicAccessRemoved;
 function tutorAssigned(owner,id){if(!owner||!id)return false;return Object.hasOwn(owner,'tutorByProgram')?owner.tutorByProgram?.music_ministry===id:owner.tutorId===id;}
 function canSee(actor,owner,assigned){
  if(!active(actor)||!active(owner)||!assigned?.includes(owner.id))return false;
  if(actor.id===owner.id)return true;
  if(['admin','departamental','secretaria','coordinador_musica'].includes(role(actor)))return true;
  if(role(actor)==='regional')return tutorAssigned(owner,actor.id)||!(Object.hasOwn(owner,'tutorByProgram')?owner.tutorByProgram?.music_ministry:owner.tutorId)&&owner.createdBy===actor.id;
  return ['administrativo','tutor'].includes(role(actor))&&(tutorAssigned(owner,actor.id)||role(actor)==='administrativo'&&owner.createdBy===actor.id);
 }
 const review=(actor,owner,assigned)=>actor?.id!==owner?.id&&cap(actor,'review')&&canSee(actor,owner,assigned);
 const final=(actor,owner,assigned)=>actor?.id!==owner?.id&&canSee(actor,owner,assigned)&&(role(actor)==='departamental'||role(actor)==='admin');
 function remove(actor,owner){
  if(!active(actor)||!active(owner)||actor.id===owner.id)return false;
  const r=role(actor),target=role(owner);
  if(r==='admin')return true;
  if(['admin','departamental'].includes(target))return false;
  if(target==='secretaria')return r==='departamental';
  return ['departamental','secretaria','coordinador_musica'].includes(r);
 }
 const api=Object.freeze({policies,cap,region,sameRegion,active,tutorAssigned,canSee,review,final,remove});
 root.SGLMusicPolicy560=api;if(typeof module!=='undefined'&&module.exports)module.exports=api;
})(typeof window==='undefined'?globalThis:window);
