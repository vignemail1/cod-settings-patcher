package main

import (
	"bytes"
	"testing"
)

func TestApplyRulesPreservesFormattingAndCRLF(t *testing.T) {
	input := []byte(
		"RendererWorkerCount@abc = 8   // commentaire  \r\n" +
			"ShowBlood@def = true\t\r\n" +
			"// TerrainQuality@ghi = High\r\n" +
			"TerrainQuality@ghi = High",
	)

	want := []byte(
		"RendererWorkerCount@abc = 10   // commentaire  \r\n" +
			"ShowBlood@def = false\t\r\n" +
			"// TerrainQuality@ghi = High\r\n" +
			"TerrainQuality@ghi = Very Low",
	)

	got, changes := applyRules(input)
	if !bytes.Equal(got, want) {
		t.Fatalf("contenu inattendu:\n got: %q\nwant: %q", got, want)
	}
	if len(changes) != 3 {
		t.Fatalf("nombre de changements = %d, want 3", len(changes))
	}
}

func TestApplyRulesIsIdempotent(t *testing.T) {
	input := []byte("ShowBlood@x = false // ok\nTerrainQuality@y = Very Low\n")
	got, changes := applyRules(input)
	if !bytes.Equal(got, input) {
		t.Fatalf("le contenu ne doit pas changer : got %q, want %q", got, input)
	}
	if len(changes) != 0 {
		t.Fatalf("nombre de changements = %d, want 0", len(changes))
	}
}

func TestApplyRulesPreservesMixedEndings(t *testing.T) {
	input := []byte("ShowBrass@x = true\r\nCorpseLimit@y = 12\nBulletImpacts@z = true\r")
	want := []byte("ShowBrass@x = false\r\nCorpseLimit@y = 0\nBulletImpacts@z = false\r")
	got, _ := applyRules(input)
	if !bytes.Equal(got, want) {
		t.Fatalf("fins de ligne non préservées:\n got: %q\nwant: %q", got, want)
	}
}

func TestApplyRulesKeepsUnknownKeysUntouched(t *testing.T) {
	input := []byte("UnknownSetting@x = keep me  // preserve\n")
	got, changes := applyRules(input)
	if !bytes.Equal(got, input) {
		t.Fatalf("une clé inconnue a été modifiée : got %q, want %q", got, input)
	}
	if len(changes) != 0 {
		t.Fatalf("nombre de changements = %d, want 0", len(changes))
	}
}
