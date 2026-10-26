'use strict';
const assert = require('node:assert/strict');
const D = require('../app/music-product-data.js');
const P = 'music_ministry';
const central = {modules:[{id:P,active:true,assignedUserIds:['learner']}],users:[]};
const local = {modules:[{id:P,active:true,assignedUserIds:['learner','new-learner']}],users:[]};
const input = {users:[{id:'coordinator',role:'coordinador_sgm',fullName:'Coordinador'}],programProgress:{
  'unassigned-admin':{[P]:{status:{'LM-PRE-1':'pending'},submissions:{'':{text:''}}}},
  learner:{[P]:{status:{'LM-PRE-1':'in_review'}}},
  'new-learner':{[P]:{status:{'LM-PRE-1':'pending'}}}
}};
const result = D.scopePatch(input,central,local);
assert.equal(result.users[0].role,'coordinador_sgm');
assert.equal(result.programProgress['unassigned-admin'],undefined);
assert.equal(result.programProgress.learner[P].status['LM-PRE-1'],'in_review');
assert(result.programProgress['new-learner']);
assert(input.programProgress['unassigned-admin']); // projection does not mutate the user's local draft
const policy = require('../app/music-policy-560.js');
for(const key of Object.keys(policy.policies.admin))assert.equal(policy.cap('coordinador_sgm',key),policy.cap('admin',key));
assert.equal(policy.remove({id:'staff',role:'secretaria',active:true},{id:'coord',role:'coordinador_sgm',active:true}),false);
console.log('Interfaz: proyección de cuentas, formación y permisos de Coordinador SGM aprobados.');
