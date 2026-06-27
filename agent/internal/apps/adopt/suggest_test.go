package adopt

import "testing"

func TestSuggestStreamlit(t *testing.T) {
	dir := mkrepo(t, map[string]string{
		"requirements.txt": "streamlit==1.58\nrequests\n",
		"webui/Main.py":    "import streamlit\n",
	})
	s := SuggestContract(dir, Detect(dir))
	if s.Contract != "http" || s.Port != 8501 {
		t.Fatalf("streamlit → http/8501, dapet %+v", s)
	}
	if len(s.StartCmd) == 0 || s.StartCmd[0] == "" {
		t.Fatalf("start_cmd kosong: %+v", s)
	}
}

func TestSuggestFastAPI(t *testing.T) {
	dir := mkrepo(t, map[string]string{
		"requirements.txt": "fastapi\nuvicorn\n",
		"main.py":          "app=1\n",
	})
	s := SuggestContract(dir, Detect(dir))
	if s.Contract != "http" || s.Port != 8000 {
		t.Fatalf("fastapi → http/8000, dapet %+v", s)
	}
}

func TestSuggestNode(t *testing.T) {
	dir := mkrepo(t, map[string]string{
		"package.json": `{"name":"x","dependencies":{"next":"14"}}`,
	})
	s := SuggestContract(dir, Detect(dir))
	if s.Contract != "http" || s.Port != 3000 {
		t.Fatalf("next → http/3000, dapet %+v", s)
	}
}

func TestSuggestPlainCLI(t *testing.T) {
	dir := mkrepo(t, map[string]string{"requirements.txt": "requests\nclick\n"})
	if s := SuggestContract(dir, Detect(dir)); s.Contract != "cli" {
		t.Fatalf("repo non-server → cli, dapet %+v", s)
	}
}
