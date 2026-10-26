'use strict';

(function installClassPolicy(root){
  const normalizeClasses = classes => Array.isArray(classes) ? classes : [];
  const assignedIds = item => Array.isArray(item?.assignedUserIds) ? item.assignedUserIds : [];
  const inactiveUser = user => !user || user.active === false || user.accountState === 'inactive' || user.accountState === 'archived';
  const classCatalogManager = user => !!user && ['admin','instructor'].includes(user.role);

  function visibleClasses(user, classes){
    const items = normalizeClasses(classes);
    if (inactiveUser(user)) return [];
    if (classCatalogManager(user)) return items;
    if (!['aspirant','pastor_aspirant'].includes(user.role)) return [];
    return items.filter(item => item.active !== false && assignedIds(item).includes(user.id));
  }

  function canAccessClasses(user, classes){
    if (inactiveUser(user)) return false;
    if (classCatalogManager(user)) return true;
    if (!['aspirant','pastor_aspirant'].includes(user.role)) return false;
    return visibleClasses(user, classes).length > 0;
  }

  function assignmentCount(userId, classes, exceptClassId=''){
    return normalizeClasses(classes).filter(item => item?.id !== exceptClassId && assignedIds(item).includes(userId)).length;
  }

  function validateAssignments(classId, selectedUserIds, classes, users, maxAssignments=3){
    const unique = [...new Set(Array.isArray(selectedUserIds) ? selectedUserIds : [])];
    const people = Array.isArray(users) ? users : [];
    for (const userId of unique){
      const person = people.find(user => user.id === userId);
      if (!person || !['aspirant','pastor_aspirant'].includes(person.role) || person.active === false || person.accountState === 'inactive' || person.accountState === 'archived') {
        return {ok:false, userIds:[], message:'Solo los Aspirantes activos pueden ser asignados como Instructores.'};
      }
      const count = assignmentCount(userId, classes, classId);
      if (count >= maxAssignments) {
        return {ok:false, userIds:[], message:`${person.fullName || 'El Aspirante'} ya tiene ${maxAssignments} clases asignadas. El maximo permitido es ${maxAssignments}.`};
      }
    }
    return {ok:true, userIds:unique, message:''};
  }

  root.ClassPolicy = Object.freeze({visibleClasses, canAccessClasses, assignmentCount, validateAssignments, classCatalogManager});
})(globalThis);
