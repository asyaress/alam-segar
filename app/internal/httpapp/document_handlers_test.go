package httpapp

import (
	"testing"
	"time"
)

func TestDocumentSignatureBindsDocumentActorAndTime(t *testing.T) {
	at := time.Date(2026, 10, 26, 1, 2, 3, 4000, time.UTC)
	one := documentSignature("rahasia", "hash-dokumen", 10, at)
	if one == "" || one != documentSignature("rahasia", "hash-dokumen", 10, at) {
		t.Fatal("tanda tangan harus deterministik untuk input yang sama")
	}
	if one == documentSignature("rahasia", "hash-berbeda", 10, at) {
		t.Fatal("tanda tangan harus berubah jika dokumen berubah")
	}
	if one == documentSignature("rahasia", "hash-dokumen", 11, at) {
		t.Fatal("tanda tangan harus berubah jika penandatangan berubah")
	}
	if one == documentSignature("rahasia", "hash-dokumen", 10, at.Add(time.Second)) {
		t.Fatal("tanda tangan harus berubah jika waktu berubah")
	}
}

func TestDocumentNeedsTenant(t *testing.T) {
	if !documentNeedsTenant("lease") || !documentNeedsTenant("inspection") {
		t.Fatal("kontrak dan serah terima harus memerlukan tanda tangan penghuni")
	}
	if documentNeedsTenant("invoice") || documentNeedsTenant("receipt") {
		t.Fatal("invoice dan kuitansi tidak boleh memerlukan tanda tangan penghuni")
	}
}
