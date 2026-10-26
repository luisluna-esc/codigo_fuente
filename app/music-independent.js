'use strict';
(()=>{
 const D=window.MusicaData,P=D.PROGRAM;let baseline=null;
 const blank=blankDatabase;blankDatabase=function(){return D.project(blank());};window.blankDatabase=blankDatabase;
 const migration=migrateDatabase;migrateDatabase=function(...args){if(database)database=D.project(database);const result=migration.apply(this,args);if(database){database=D.project(database);v3ActiveProgramId=P;}return result;};window.migrateDatabase=migrateDatabase;
 const saving=persist;persist=async function(...args){if(args[0]!==false&&navigator.onLine===false)throw Error('Necesitas conexión con el servidor para guardar.');if(database)database=D.project(database);return saving.apply(this,args);};window.persist=persist;
 const api=v3Api;v3Api=async function(path,options={}){const next={...options};if(options.body){next.body=structuredClone(options.body);if(next.body.programId&&next.body.programId!==P)throw Error('Solo se permite Ministerio de la Música.');if(next.body.database)next.body.database=D.project(next.body.database);if(next.body.patch)next.body.patch=D.scopePatch(next.body.patch,baseline||database||{},database||{},baseline||D.project(database));}return api.call(this,path,next);};window.v3Api=v3Api;
 const apply=v3ApplyServerSnapshot;v3ApplyServerSnapshot=async function(snapshot,...args){const result=await apply.call(this,D.snapshot(snapshot),...args);if(snapshot?.database&&!result?.deferred&&!result?.stale&&!result?.blocked)baseline=D.project(snapshot.database);return result;};window.v3ApplyServerSnapshot=v3ApplyServerSnapshot;
 appLogo=function(){return database?.settings?.musicAppLogo||DEFAULT_APP_LOGO;};window.appLogo=appLogo;
 const assigned=v3AssignedProgramIds;v3AssignedProgramIds=function(user){return assigned(user).filter(id=>id===P);};window.v3AssignedProgramIds=v3AssignedProgramIds;
 const has=v3UserHasProgram;v3UserHasProgram=function(user,id){return id===P&&has(user,id);};window.v3UserHasProgram=v3UserHasProgram;
 const forbidden=page=>['classes','clubs','leadership','drive','conflicts'].includes(page)||String(page).startsWith('class_')||(String(page).startsWith('module_')&&page!=='module_music_ministry');
 const navigation=navItems;navItems=function(user){return navigation(user).filter(x=>!forbidden(x[0])&&x[0]!=='module_music_ministry');};window.navItems=navItems;
 const navigate=go;go=function(page,...args){v3ActiveProgramId=P;if(forbidden(page))return toast('Esta opción no forma parte de Ministerio de la Música.');return navigate.call(this,page,...args);};window.go=go;
 renderDashboard=function(){v3ActiveProgramId=P;if(v3UserHasProgram(currentUser(),P))return SGL550.open('home');setTop('Ministerio de la Música','Inicio');setContent('<article class="card"><h1>Ministerio de la Música</h1><p>Solicita al Administrador que habilite tu formación en Música.</p></article>');};window.renderDashboard=renderDashboard;
 window.MusicaApp={};
})();
