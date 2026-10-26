package main

import (
 "net/http/httptest"
 "os"
 "os/exec"
 "testing"
 "time"
)

func Test594MusicTutorAllTrainingRoles(t *testing.T) {
 for _, role := range []string{"departamental","secretaria","administrativo","coordinador_musica","regional","tutor","aspirant","pastor_aspirant"} {
  t.Run(role,func(t *testing.T){
   s,h:=setup555(t)
   st,_:=s.loadLegacy(); db:=legacyMap(st.State.Database); u:=legacyUser(db,"administrativo");u["role"]=role
   u["tutorByProgram"]=map[string]any{musicProgram553:"tm"}
   if !musicTutorReady571(db,u){t.Fatal("valid tutor rejected")}
   if e:=s.saveLegacy(st);e!=nil{t.Fatal(e)}
   _,login:=musicLogin553(t,h,"administrativo",true);token:=stringValue(login["sessionToken"])
   code,snap:=call555(t,h,"GET",musicAPI553+"/api/accounts/view?actorId=administrativo",token,true,nil)
   if code!=200{t.Fatal(code,snap)}
   source:=legacyMap(snap["database"]);p:=reviewProgressContainer(source,"administrativo",musicProgram553)
   ensureMapField(p,"status")["LM-PRE-2"]="in_progress"
   ensureMapField(p,"submissions")["LM-PRE-2"]=map[string]any{"text":"Avance personal de prueba.\nSegundo párrafo de prueba.\nTercer párrafo de prueba.","attachmentIds":[]any{},"updatedAt":time.Now().UTC().Format(time.RFC3339Nano)}
   code,result:=call555(t,h,"POST",musicAPI553+"/api/sync",token,true,map[string]any{"actorId":"administrativo","actorRole":"aspirant","installationId":"test594","baseVersion":snap["serverVersion"],"restoreGeneration":1,"operationId":"tutor-594-"+role,"operationType":"save_progress","programId":musicProgram553,"ownerId":"administrativo","requirementId":"LM-PRE-2","database":source,"files":[]any{},"ackOnly":true})
   if code!=200{t.Fatal(code,result)}
   st,_=s.loadLegacy();db=legacyMap(st.State.Database)
   if stringValue(mapField(legacyProgressForRequirement(db,"administrativo",musicProgram553,false),"status")["LM-PRE-2"])!="in_progress"{t.Fatal("progress not persisted")}
   u=legacyUser(db,"administrativo");if stringValue(u["role"])!=role{t.Fatal("authority role changed")}
   for _, id:=range []string{"","missing","administrativo","tg"}{u["tutorByProgram"]=map[string]any{musicProgram553:id};if musicTutorReady571(db,u){t.Fatal("invalid tutor accepted",id)}}
   u["tutorByProgram"]=map[string]any{musicProgram553:"tm"};legacyUser(db,"tm")["accountState"]="inactive";if musicTutorReady571(db,u){t.Fatal("inactive tutor accepted")}
  })
 }
}

func Test594Browser(t *testing.T){
 if os.Getenv("SGL594_BROWSER")!="1"{t.Skip("browser opt-in")}
 _,h:=setup555(t);server:=httptest.NewServer(h);defer server.Close()
 cmd:=exec.Command("node","../../../tests/browser-tutor-594.cjs");cmd.Env=append(os.Environ(),"SGL_BROWSER_BASE="+server.URL)
 out,e:=cmd.CombinedOutput();t.Log(string(out));if e!=nil{t.Fatal(e)}
}
