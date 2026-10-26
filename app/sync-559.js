/* Shared delivery and reception for every SGL module. Durable queues stay intact. */
'use strict';
(()=>{
 let timer=null,running=null;
 const actor=()=>window.SGL555?.actual()||currentUser();
 const ownPending=()=>sgl559OwnPending(database?.syncQueue,actor()?.id);
 function ready(){return !!actor()&&!document.hidden&&!window.SGL555?.isSwitching?.()&&!!database?.settings?.serverUrl&&v3CanAttemptNetwork();}
 function schedule(delay=180){clearTimeout(timer);timer=setTimeout(()=>{timer=null;pump().catch(()=>{});},delay);}
 async function pump(){
  if(running)return running;
  if(!ready())return false;
  const id=actor().id;
  running=(async()=>{
   try{
    if(database.settings?.autoSync!==false){
     let rounds=0;
     while(actor()?.id===id&&ready()&&rounds++<12){
      const op=sgl559NextOperation(database.syncQueue,id,actor().role==='admin'&&!SGL555.personalMode()?'admin_change':'');
      if(!op)break;
      const before=JSON.stringify([op.id,op.status,op.attempts,op.nextAttemptAt]);
      await syncNow();
      const live=(database.syncQueue||[]).find(x=>x.id===op.id);
      if(live&&before===JSON.stringify([live.id,live.status,live.attempts,live.nextAttemptAt]))break;
     }
    }
   }finally{
    if(actor()?.id===id&&ready())await window.SGL556?.refreshRoster();
   }
   return ownPending().length===0;
  })().finally(()=>{running=null;});
  return running;
 }
 const basePersist=persist;
 persist=async function(markPending=true,...args){
  const result=await basePersist.call(this,markPending,...args);
  if(markPending&&database?.settings?.autoSync!==false)schedule();
  return result;
 };window.persist=persist;
 const baseSaveUser=saveUser;
 saveUser=async function(event,...args){
  const id=actor()?.id,result=await baseSaveUser.call(this,event,...args);
  if(actor()?.id!==id||document.getElementById('user_error'))return result;
  if(ready()&&actor()?.role==='admin'&&ownPending().some(op=>op.action==='admin_change')){
   await syncNow();
   const pending=ownPending().filter(op=>op.action==='admin_change');
   toast(pending.length?'Guardado en este equipo · envío pendiente al servidor'+(pending[0].lastError?': '+pending[0].lastError:''):'Usuario y accesos confirmados en el servidor.');
  }
  schedule();return result;
 };window.saveUser=saveUser;
 for(const name of ['focus','online','pageshow'])window.addEventListener(name,()=>schedule(80));
 document.addEventListener('visibilitychange',()=>{if(!document.hidden)schedule(80);});
 setInterval(()=>{if(ready()&&ownPending().length)schedule(80);},10000);
 window.SGL559={pump,schedule,ownPending};
 schedule(500);
})();
