'use strict';
(()=>{
 const MUSIC='music_ministry',YEAR='LM-NUT-1',ROLES=['departamental','administrativo','secretaria','regional'];
 let deleting=false,scheduled=false;
 function canRemove(ownerId){
  const actor=currentUser(),owner=userById(ownerId);
  if(!actor||!owner||SGL555.personalMode()||actor.id===ownerId||!ROLES.includes(actor.role)||!SGL555.managementPrograms(SGL555.actual()).includes(MUSIC)||!SGL555.canSee(actor,owner,MUSIC))return false;
  return v3SubmissionFor(ownerId,YEAR,MUSIC).bibleReadings?.choice==='reavivados';
 }
 async function confirmReavivados(ownerId){
  if(deleting||!canRemove(ownerId))return toast('No tienes permiso para eliminar este avance.');
  if(!await SGL531.beforeLeave())return;
  const actor=currentUser().id,owner=userById(ownerId),progress=v3EnsureProgramProgress(ownerId,MUSIC);
  const op={actorId:actor,ownerId,programId:MUSIC,requirementId:YEAR,scope:'requirement',operationId:newId('progress_reset'),baseUpdatedAt:progress.updatedAt||'',restoreGeneration:database.restoreGeneration||1};
  showModal(`<div class="modal small"><div class="modal-head"><h3>Eliminar avance de Reavivados por su Palabra</h3><button class="modal-close" onclick="closeModal()">×</button></div><div class="modal-body"><p><strong>${esc(owner.fullName)}</strong></p><p>Se eliminarán las lecturas, el informe y la validación de este requisito. Quedará Pendiente y necesitará una nueva revisión. Los demás requisitos se conservan.</p><p id="sgl557_reset_error" class="error" role="alert"></p></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cancelar</button><button class="btn red" id="sgl557_confirm_reset">Eliminar avance definitivamente</button></div></div>`);
  document.getElementById('sgl557_confirm_reset').onclick=async()=>{
   if(deleting)return;const button=document.getElementById('sgl557_confirm_reset');deleting=true;button.disabled=true;
   const epoch=Number(window.__sgl555NavigationEpoch||0),view=SGL555.personalMode();
   try{
    const {payload}=await v3Api('/api/progress/reset',{method:'POST',body:op,timeout:20000});
    if(currentUser()?.id!==actor||SGL555.personalMode()!==view)return;
    await v3ApplyServerSnapshot(payload);
    if(epoch===Number(window.__sgl555NavigationEpoch||0)){await closeModal();if(currentPage==='requirements')renderRequirements();else render();}
    toast('Avance eliminado. Reavivados quedó Pendiente y requiere nueva revisión.');
   }catch(error){const box=document.getElementById('sgl557_reset_error');if(box)box.textContent=error.message+' No se eliminó el avance en este equipo.';}
   finally{deleting=false;if(button.isConnected)button.disabled=false;}
  };
 }
 function control(host,ownerId){
  if(!host||!canRemove(ownerId)||host.querySelector('[data-reavivados-reset557]'))return;
  const button=document.createElement('button');button.type='button';button.className='btn red small';button.dataset.reavivadosReset557=ownerId;button.textContent='Eliminar avance de Reavivados';button.onclick=()=>confirmReavivados(ownerId);host.append(button);
 }
 function enhance(){
  for(const b of document.querySelectorAll('[data-reavivados-reset557]'))if(!canRemove(b.dataset.reavivadosReset557))b.remove();
  if(v3ActiveProgramId===MUSIC&&currentPage==='requirements'){
   const owner=targetAspirant();for(const row of document.querySelectorAll(`[data-sgl5059-requirement="${YEAR}"],[data-sgl5058-requirement="${YEAR}"]`))control(row.querySelector('.req-actions,.actions'),owner?.id);
  }
  if(v3ModalProgramId===MUSIC&&modalRequirementId===YEAR&&!document.getElementById('sgl557_confirm_reset'))control(document.querySelector('#modal_root .modal-foot'),modalOwnerId);
  if(currentPage==='music_module'&&SGL550.getSection()==='readings'){
   const owner=document.getElementById('sgl550_owner')?.value||targetAspirant()?.id;
   const card=[...document.querySelectorAll('.sgl551-reading-options article')].find(x=>x.querySelector('h3')?.textContent==='Reavivados por su Palabra');control(card,owner);
  }
 }
 new MutationObserver(()=>{if(scheduled)return;scheduled=true;requestAnimationFrame(()=>{scheduled=false;enhance();});}).observe(document.body,{subtree:true,childList:true});
 window.SGL557={canRemove,confirmReavivados,enhance};
})();
