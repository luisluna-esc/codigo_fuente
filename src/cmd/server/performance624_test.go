package main
import("testing";"os";"encoding/json";"net/http/httptest";"time";"sort";"fmt")
func TestPerformance624(t *testing.T){
 path:=os.Getenv("MUSIC_PERF_BACKUP");if path==""{t.Skip("local benchmark")};b,e:=os.ReadFile(path);if e!=nil{t.Fatal(e)};var backup map[string]any;if json.Unmarshal(b,&backup)!=nil{t.Fatal("backup")}
 s,db:=fixture550(t);musicEnsure610(db);d:=musicData550(db);d["templates"]=cloneAny(mapField(mapField(backup,"database"),"music550")["templates"])
 st,_:=s.loadLegacy();st.State.Database=db;s.saveLegacy(st)
 for _,resource:=range []string{"templates","certificates","followup"}{
  samples:=[]float64{};size:=0
  for i:=0;i<13;i++{r:=httptest.NewRequest("GET","/api/music/v610?actorId=admin&resource="+resource,nil);r.Header.Set("Authorization","Bearer admin");w:=httptest.NewRecorder();start:=time.Now();s.music610(w,r);elapsed:=float64(time.Since(start).Microseconds())/1000;if w.Code!=200{t.Fatal(w.Code,w.Body.String())};if i>=3{samples=append(samples,elapsed)};size=w.Body.Len()}
  sort.Float64s(samples);fmt.Printf("PERF624 resource=%s median_ms=%.3f p95_ms=%.3f bytes=%d samples=%d\n",resource,samples[len(samples)/2],samples[len(samples)-1],size,len(samples))
 }
}
