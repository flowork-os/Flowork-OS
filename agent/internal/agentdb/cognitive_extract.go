// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

var ValidNodeTypes = map[string]bool{
	"person": true, "concept": true, "project": true, "trait": true, "event": true,
	"skill": true, "fact": true, "preference": true, "doctrine": true, "persona": true,
	"memory": true, "knowledge": true,
}

type ExtractedNode struct {
	Label       string  `json:"label"`
	Type        string  `json:"type"`
	Why         string  `json:"why"`
	Who         string  `json:"who"`
	WhereDomain string  `json:"where_domain"`
	WhenValid   string  `json:"when_valid"`
	SourceKind  string  `json:"source_kind"`
	Confidence  float64 `json:"confidence"`
}

type ExtractedEdge struct {
	FromLabel    string  `json:"from_label"`
	ToLabel      string  `json:"to_label"`
	RelationType string  `json:"relation_type"`
	SourceKind   string  `json:"source_kind"`
	Confidence   float64 `json:"confidence"`
}

type ExtractResult struct {
	Nodes   []ExtractedNode
	Edges   []ExtractedEdge
	Dropped []string
}

func BuildExtractPrompt(conversation string) string {
	rels := sortedKeys(ValidRelations)
	types := sortedKeys(ValidNodeTypes)
	var b strings.Builder
	b.WriteString("You distill a conversation into a small knowledge graph. Output STRICT JSON only.\n\n")
	b.WriteString("RULES:\n")
	b.WriteString("- Extract ONLY what still matters a month from now. Drop chit-chat (greetings, 'ok', 'thanks').\n")
	b.WriteString("- Separate FACTS (type fact/project/concept) from the USER'S TRAITS (type trait/preference).\n")
	b.WriteString("- relation_type MUST be one of: " + strings.Join(rels, ", ") + ". If none fits, omit the edge.\n")
	b.WriteString("- node type MUST be one of: " + strings.Join(types, ", ") + ".\n")
	b.WriteString("- source_kind: 'user_said' if the user stated it directly, else 'agent_inferred'.\n")
	b.WriteString("- NEVER invent facts, numbers, or sources. confidence in [0,1].\n")
	b.WriteString("- Edges reference nodes by their label (from_label/to_label).\n\n")
	b.WriteString("OUTPUT SHAPE:\n")
	b.WriteString(`{"nodes":[{"label","type","why","who","where_domain","when_valid","source_kind","confidence"}],`)
	b.WriteString(`"edges":[{"from_label","to_label","relation_type","source_kind","confidence"}]}` + "\n\n")
	b.WriteString("CONVERSATION:\n")
	b.WriteString(conversation)
	return b.String()
}

func ParseExtraction(raw string) (ExtractResult, error) {
	s := stripCodeFence(strings.TrimSpace(raw))
	var doc struct {
		Nodes []ExtractedNode `json:"nodes"`
		Edges []ExtractedEdge `json:"edges"`
	}
	if err := json.Unmarshal([]byte(s), &doc); err != nil {
		return ExtractResult{}, fmt.Errorf("parse extraction JSON: %w", err)
	}

	var res ExtractResult
	for _, n := range doc.Nodes {
		n.Label = strings.TrimSpace(n.Label)
		n.Type = strings.TrimSpace(strings.ToLower(n.Type))
		if n.Label == "" || n.Type == "" {
			res.Dropped = append(res.Dropped, "node: label/type kosong")
			continue
		}
		if !ValidNodeTypes[n.Type] {
			res.Dropped = append(res.Dropped, "node: type invalid '"+n.Type+"'")
			continue
		}
		n.SourceKind = normSourceKind(n.SourceKind)
		n.Confidence = clamp01(n.Confidence)
		if len(n.Label) > maxCogLabelBytes {
			n.Label = n.Label[:maxCogLabelBytes]
		}
		res.Nodes = append(res.Nodes, n)
	}
	for _, e := range doc.Edges {
		e.FromLabel = strings.TrimSpace(e.FromLabel)
		e.ToLabel = strings.TrimSpace(e.ToLabel)
		e.RelationType = strings.TrimSpace(strings.ToLower(e.RelationType))
		if e.FromLabel == "" || e.ToLabel == "" || e.RelationType == "" {
			res.Dropped = append(res.Dropped, "edge: field kosong")
			continue
		}
		if !ValidRelations[e.RelationType] {
			res.Dropped = append(res.Dropped, "edge: relation invalid '"+e.RelationType+"'")
			continue
		}

		if ValidRelations[strings.ToLower(e.FromLabel)] || ValidRelations[strings.ToLower(e.ToLabel)] {
			res.Dropped = append(res.Dropped, "edge: endpoint adalah kata-relasi (malformed)")
			continue
		}
		if strings.EqualFold(e.FromLabel, e.ToLabel) {
			res.Dropped = append(res.Dropped, "edge: self-loop (from==to)")
			continue
		}
		e.SourceKind = normSourceKind(e.SourceKind)
		e.Confidence = clamp01(e.Confidence)
		res.Edges = append(res.Edges, e)
	}
	return res, nil
}

func stripCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}

	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func normSourceKind(k string) string {
	k = strings.TrimSpace(strings.ToLower(k))
	switch k {
	case "user_said", "verified", "strong_model_unverified":
		return k
	default:
		return "agent_inferred"
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	if v == 0 {
		return 0.5
	}
	return v
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
