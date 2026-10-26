'use strict';
(function(root){
 const MUSIC='music_ministry',MC=['class_grouped','guide_major','guide_master','guide_master_advanced'];
 const program=(id,view)=>view==='mmus'?id===MUSIC:MC.includes(id);
 function person(u,view,personal=[],management=[],hasClass=false){return !!u&&(u.role==='admin'||(view==='mmus'&&u.musicAccessRemoved!==true||view==='mc')&&[...personal,...management].some(p=>program(p,view))||view==='mc'&&hasClass);}
 function nav(items,view){return items.filter(([id])=>{
  if(id.startsWith('module_'))return program(id.slice(7),view);
  if(view==='mc')return !id.startsWith('music_');
  return !['classes','clubs','leadership'].includes(id)&&!id.startsWith('class_');
 });}
 function birthdays(users,actor,db,view,now=new Date()){
  const region=x=>String(x||'').trim().toUpperCase().replace(/^REGI[ÓO]N\s*/,'').replace(/^0+(?=\d)/,'');
  const staff=['admin','departamental','secretaria','coordinador_musica','administrativo'];
  const member=(u,music)=>!(music&&u.musicAccessRemoved===true)&&[...MC,MUSIC].filter(p=>(p===MUSIC)===music).some(p=>(db.modules||[]).some(m=>m.id===p&&m.active!==false&&(m.assignedUserIds||[]).includes(u.id))||(u.role==='admin'||u.role==='tutor'?(u.role==='admin'||(u.tutorProgramIds||u.managementProgramIds||[]).includes(p)):(u.managementProgramIds||[]).includes(p)));
  return users.filter(u=>u.active!==false&&(!u.accountState||u.accountState==='active')&&/^\d{4}-\d{2}-\d{2}$/.test(u.birthDate||'')).filter(u=>{
   const music=member(u,true);if(view==='mmus'?!music:!member(u,false))return false;
   const assigned=('tutorByProgram' in u?Object.values(u.tutorByProgram||{}).includes(actor?.id):u.tutorId===actor?.id);
   const today=u.birthDate.slice(5)===String(now.getMonth()+1).padStart(2,'0')+'-'+String(now.getDate()).padStart(2,'0');
   return view==='mmus'&&music&&(root.SGL602?.permit('birthdays.all')??db.music550?.effective602?.['birthdays.all']??true)||actor?.id===u.id||staff.includes(actor?.role)||actor?.role==='regional'&&((region(actor.region)&&region(actor.region)===region(u.region))||(actor.regionalAssignedUserIds||[]).includes(u.id)||Object.values(u.regionalByProgram||{}).includes(actor.id)||assigned)||actor?.role==='tutor'&&assigned||music&&today;
  }).map(u=>({id:u.id,fullName:u.fullName,birthDate:'2000-'+u.birthDate.slice(5)}));
 }
 const api=Object.freeze({MUSIC,MC,program,person,nav,birthdays});root.SGLViewPolicy562=api;if(typeof module!=='undefined')module.exports=api;
})(typeof window==='undefined'?globalThis:window);
