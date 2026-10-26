package main
import("encoding/json";"os";"testing")
// Opt-in read-only audit. Never log credentials or report contents.
func Test596ProvidedBackup(t *testing.T){
 path:=os.Getenv("SGL_AUDIT_BACKUP");if path==""{t.Skip("requires private decrypted backup")}
 b,e:=os.ReadFile(path);if e!=nil{t.Fatal(e)}
 var backup struct{Database map[string]any `json:"database"`;Files []any `json:"files"`}
 if e=json.Unmarshal(b,&backup);e!=nil{t.Fatal(e)}
 db:=backup.Database;total,valid:=0,0
 for _,raw:=range arrayField(db,"users"){u:=legacyMap(raw);id:=stringValue(u["id"]);if assignedProgram555(db,id,musicProgram553){total++;if musicTutorReady571(db,u){valid++}else{t.Logf("INVALID_MUSIC_TUTOR id=%s",id)}}}
 t.Logf("MUSIC_TUTORS assigned=%d valid=%d",total,valid)
 for _,raw:=range arrayField(db,"modules"){m:=legacyMap(raw);pid:=stringValue(m["id"]);for _,rid:=range arrayField(m,"assignedUserIds"){id:=stringValue(rid);u:=legacyUser(db,id);tid:=legacyTutorForProgram(u,pid);tutor:=legacyUser(db,tid);if tid==""||tid==id||!legacyActive(tutor)||!managementProgram555(db,tutor,pid){t.Logf("ASSIGNMENT_NEEDS_REVIEW program=%s id=%s",pid,id)}}}
 before:=map[string]any{};for _,k:=range []string{"users","progress","programProgress","modules"}{before[k]=db[k]};snapshot,_:=json.Marshal(before);files,_:=json.Marshal(backup.Files)
 st:=legacyStore{State:legacyState{Database:db,Files:backup.Files,RestoreGeneration:2}};normalizeLegacyStore(&st)
 after:=map[string]any{};for k:=range before{after[k]=legacyMap(st.State.Database)[k]};a,_:=json.Marshal(after);f,_:=json.Marshal(st.State.Files)
 if string(snapshot)!=string(a)||string(files)!=string(f){t.Fatal("normalization changed stored work")}
 t.Log("NORMALIZATION preserves users, modules, all progress and files")
}
