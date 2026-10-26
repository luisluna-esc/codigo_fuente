'use strict';
// Explicit tutor assignments are independent of whether an aspirant already exists.
function sgl548TutorPrograms(person){
 const active=(database.modules||[]).filter(m=>m.active!==false);
 if(Array.isArray(person?.tutorProgramIds))return active.filter(m=>person.tutorProgramIds.includes(m.id)).map(m=>m.id);
 const ids=active.filter(m=>(m.assignedUserIds||[]).some(id=>sgl545HasTutor(userById(id),person?.id,m.id))).map(m=>m.id);
 return ids.length?ids:active.filter(m=>m.id==='class_grouped').map(m=>m.id);
}
window.sgl548TutorPrograms=sgl548TutorPrograms;
(()=>{
 const assigned=v3AssignedProgramIds;
 v3AssignedProgramIds=function(person){return person?.role==='tutor'&&sgl421AccountActive(person)?sgl548TutorPrograms(person):assigned(person);};window.v3AssignedProgramIds=v3AssignedProgramIds;
 const open=openUser;
 openUser=function(id=''){
  const result=open.apply(this,arguments),actor=currentUser(),person=id?userById(id):null,form=document.querySelector('.user-editor .field-row');
  if(!form||!['admin','regional','tutor'].includes(actor?.role))return result;
  if(actor.role==='tutor'){
   const group=document.createElement('div');group.id='user_module_assignments';group.className='field-group full user-module-assignment';
   group.innerHTML='<span>Módulos que cursará el Aspirante</span><div class="module-checkbox-grid">'+sgl548TutorPrograms(actor).map((p,i)=>`<label class="choice"><input type="checkbox" name="user_module" value="${esc(p)}" ${i===0?'checked':''}><span>${esc(v3Program(p).menu_label||v3Program(p).name)}</span></label>`).join('')+'</div><small>Solo puedes asignar los módulos habilitados para tu tutoría.</small>';form.append(group);return result;
  }
  const group=document.createElement('div');group.id='sgl548_tutor_programs';group.className='field-group full';
  const programs=(database.modules||[]).filter(m=>m.active!==false&&(actor.role!=='regional'||REGIONAL_PROGRAM_IDS.includes(m.id)));
  const current=person?sgl548TutorPrograms(person):[];
  group.innerHTML='<span>Módulos en los que será Tutor</span><div class="module-checkbox-grid">'+programs.map(m=>`<label class="choice"><input type="checkbox" name="sgl548_tutor_program" value="${esc(m.id)}" ${current.includes(m.id)?'checked':''}><span>${esc(v3Program(m.id).menu_label||v3Program(m.id).name)}</span></label>`).join('')+'</div>';
  form.append(group);const toggle=()=>{group.hidden=!['tutor','lider_musica'].includes(value('user_role'));};document.getElementById('user_role')?.addEventListener('change',toggle);toggle();
  // Restrict individual tutor pickers to the configured program, while retaining inactive historic values for editing.
  for(const field of form.querySelectorAll('[data-sgl545-program]'))for(const option of [...field.options])if(option.value&&!option.selected&&!sgl548TutorPrograms(userById(option.value)).includes(field.dataset.sgl545Program))option.remove();
  return result;
 };window.openUser=openUser;
 const save=saveUser;
 saveUser=async function(event,id=''){
  if(['tutor','lider_musica'].includes(value('user_role'))&&document.getElementById('sgl548_tutor_programs')&&!document.querySelector('[name=sgl548_tutor_program]:checked')){event.preventDefault();return setText('user_error','Selecciona al menos un módulo para el Tutor.');}
  return save.apply(this,arguments);
 };window.saveUser=saveUser;
})();
