package main

import (
	"bytes"
	"fmt"
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

	got, changes := applyRules(input, testValues())
	if !bytes.Equal(got, want) {
		t.Fatalf("contenu inattendu:\n got: %q\nwant: %q", got, want)
	}
	if len(changes) != 3 {
		t.Fatalf("nombre de changements = %d, want 3", len(changes))
	}
}

func TestApplyRulesIsIdempotent(t *testing.T) {
	input := []byte("ShowBlood@x = false // ok\nTerrainQuality@y = Very Low\n")
	got, changes := applyRules(input, testValues())
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
	got, _ := applyRules(input, testValues())
	if !bytes.Equal(got, want) {
		t.Fatalf("fins de ligne non préservées:\n got: %q\nwant: %q", got, want)
	}
}

func TestApplyRulesKeepsUnknownKeysUntouched(t *testing.T) {
	input := []byte("UnknownSetting@x = keep me  // preserve\n")
	got, changes := applyRules(input, testValues())
	if !bytes.Equal(got, input) {
		t.Fatalf("une clé inconnue a été modifiée : got %q, want %q", got, input)
	}
	if len(changes) != 0 {
		t.Fatalf("nombre de changements = %d, want 0", len(changes))
	}
}

func TestRendererWorkerCount(t *testing.T) {
	tests := []struct {
		name      string
		coreCount int
		want      int
	}{
		{name: "two cores reserve one", coreCount: 2, want: 1},
		{name: "four cores capped at cores minus one", coreCount: 4, want: 3},
		{name: "six physical cores", coreCount: 6, want: 5},
		{name: "eight P-cores", coreCount: 8, want: 7},
		{name: "twelve physical cores", coreCount: 12, want: 10},
		{name: "twenty P-cores", coreCount: 20, want: 16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := rendererWorkerCount(test.coreCount)
			if err != nil {
				t.Fatalf(
					"rendererWorkerCount(%d) a retourné une erreur : %v",
					test.coreCount,
					err,
				)
			}

			if got != test.want {
				t.Fatalf(
					"rendererWorkerCount(%d) = %d, want %d",
					test.coreCount,
					got,
					test.want,
				)
			}
		})
	}
}

func TestRendererWorkerCountRejectsInsufficientCoreCount(t *testing.T) {
	for _, coreCount := range []int{-1, 0, 1} {
		t.Run(fmt.Sprintf("%d core(s)", coreCount), func(t *testing.T) {
			if _, err := rendererWorkerCount(coreCount); err == nil {
				t.Fatalf(
					"rendererWorkerCount(%d) devait retourner une erreur",
					coreCount,
				)
			}
		})
	}
}

func testValues() map[string]string {
	values := make(map[string]string, len(desiredValues)+1)
	for key, value := range desiredValues {
		values[key] = value
	}
	values[rendererWorkerCountKey] = "10"
	return values
}
