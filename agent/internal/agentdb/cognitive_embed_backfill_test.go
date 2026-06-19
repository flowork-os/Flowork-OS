package agentdb

import ("path/filepath";"testing")

func TestEmbedBackfill_ListAndSet(t *testing.T){
	s,err:=Open(filepath.Join(t.TempDir(),"state.db")); if err!=nil{t.Fatal(err)}; defer s.Close()
	s.UpsertNode(CogNode{ID:"agent:t/tool/x",Label:"x",Type:"tool",
		Properties:`{"needs_embedding":true}`,Status:"active",SourceKind:"verified"})
	// belum embed → muncul di list
	tg,_:=s.NodesNeedingEmbedding(10)
	if len(tg)!=1 || tg[0].ID!="agent:t/tool/x"{ t.Fatalf("need-embed list = %+v",tg) }
	// set embedding → ilang dari list + flag flip
	if err:=s.SetNodeEmbedding("agent:t/tool/x",Quantize([]float32{0.1,0.2,0.3}));err!=nil{t.Fatal(err)}
	tg2,_:=s.NodesNeedingEmbedding(10)
	if len(tg2)!=0{ t.Fatalf("after set, still needs embedding: %+v",tg2) }
	n,_,_:=s.GetNode("agent:t/tool/x")
	if len(n.Embedding)==0{ t.Fatal("embedding not persisted") }
}
