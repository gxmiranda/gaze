package taxonomy

import (
	"encoding/json"
	"testing"
)

func TestGenerateID_Deterministic(t *testing.T) {
	id1 := GenerateID("pkg/foo", "Save", "ReceiverMutation", "foo.go:10:2")
	id2 := GenerateID("pkg/foo", "Save", "ReceiverMutation", "foo.go:10:2")

	if id1 != id2 {
		t.Errorf("GenerateID not deterministic: %q != %q", id1, id2)
	}
}

func TestGenerateID_Format(t *testing.T) {
	id := GenerateID("pkg/foo", "Save", "ReceiverMutation", "foo.go:10:2")

	if len(id) != 11 { // "se-" + 8 hex chars
		t.Errorf("expected ID length 11, got %d: %q", len(id), id)
	}
	if id[:3] != "se-" {
		t.Errorf("expected ID to start with 'se-', got %q", id)
	}
}

func TestGenerateID_UniqueForDifferentInputs(t *testing.T) {
	id1 := GenerateID("pkg/foo", "Save", "ReceiverMutation", "foo.go:10:2")
	id2 := GenerateID("pkg/foo", "Save", "ReturnValue", "foo.go:10:2")
	id3 := GenerateID("pkg/foo", "Load", "ReceiverMutation", "foo.go:20:2")

	if id1 == id2 {
		t.Errorf("different effect types should produce different IDs")
	}
	if id1 == id3 {
		t.Errorf("different functions should produce different IDs")
	}
}

func TestTierOf_P0Types(t *testing.T) {
	p0Types := []SideEffectType{
		ReturnValue, ErrorReturn, SentinelError,
		ReceiverMutation, PointerArgMutation,
	}
	for _, st := range p0Types {
		if got := TierOf(st); got != TierP0 {
			t.Errorf("TierOf(%s) = %s, want P0", st, got)
		}
	}
}

func TestTierOf_AllTypesHaveTiers(t *testing.T) {
	allTypes := []SideEffectType{
		// P0
		ReturnValue, ErrorReturn, SentinelError,
		ReceiverMutation, PointerArgMutation,
		// P1
		SliceMutation, MapMutation, GlobalMutation,
		WriterOutput, HTTPResponseWrite, ChannelSend,
		ChannelClose, DeferredReturnMutation,
		// P2
		FileSystemWrite, FileSystemDelete, FileSystemMeta,
		DatabaseWrite, DatabaseTransaction, GoroutineSpawn,
		Panic, CallbackInvocation, LogWrite, ContextCancellation,
		// P3
		StdoutWrite, StderrWrite, EnvVarMutation,
		MutexOp, WaitGroupOp, AtomicOp, TimeDependency,
		ProcessExit, RecoverBehavior,
		// P4
		ReflectionMutation, UnsafeMutation, CgoCall,
		FinalizerRegistration, SyncPoolOp,
		ClosureCaptureMutation,
	}

	for _, st := range allTypes {
		tier := TierOf(st)
		if tier == "" {
			t.Errorf("TierOf(%s) returned empty tier", st)
		}
	}
}

func TestTierOf_UnknownDefaultsToP4(t *testing.T) {
	if got := TierOf("UnknownType"); got != TierP4 {
		t.Errorf("TierOf(unknown) = %s, want P4", got)
	}
}

func TestClassification_JSONSerialization(t *testing.T) {
	c := Classification{
		Label:      Contractual,
		Confidence: 87,
		Signals: []Signal{
			{
				Source:     "interface",
				Weight:     30,
				SourceFile: "store.go",
				Excerpt:    "implements io.Reader",
				Reasoning:  "method satisfies interface contract",
			},
			{
				Source: "caller",
				Weight: 12,
			},
		},
		Reasoning: "strong contractual evidence",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal Classification: %v", err)
	}

	var got Classification
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal Classification: %v", err)
	}

	if got.Label != Contractual {
		t.Errorf("label = %q, want %q", got.Label, Contractual)
	}
	if got.Confidence != 87 {
		t.Errorf("confidence = %d, want 87", got.Confidence)
	}
	if len(got.Signals) != 2 {
		t.Fatalf("signals count = %d, want 2", len(got.Signals))
	}
	if got.Signals[0].Source != "interface" {
		t.Errorf("signal[0].source = %q, want %q",
			got.Signals[0].Source, "interface")
	}
	if got.Signals[0].Weight != 30 {
		t.Errorf("signal[0].weight = %d, want 30",
			got.Signals[0].Weight)
	}
}

func TestSignal_OmitEmptyFields(t *testing.T) {
	// Non-verbose signal: only source and weight populated.
	s := Signal{
		Source: "naming",
		Weight: 10,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal Signal: %v", err)
	}

	str := string(data)
	// source_file, excerpt, reasoning should be omitted.
	if contains(str, "source_file") {
		t.Errorf("non-verbose signal should omit source_file: %s", str)
	}
	if contains(str, "excerpt") {
		t.Errorf("non-verbose signal should omit excerpt: %s", str)
	}
	if contains(str, "reasoning") {
		t.Errorf("non-verbose signal should omit reasoning: %s", str)
	}
}

func TestSideEffect_WithoutClassification(t *testing.T) {
	se := SideEffect{
		ID:          "se-a1b2c3d4",
		Type:        ReturnValue,
		Tier:        TierP0,
		Location:    "foo.go:10:1",
		Description: "returns int",
		Target:      "int",
	}

	data, err := json.Marshal(se)
	if err != nil {
		t.Fatalf("Marshal SideEffect: %v", err)
	}

	// classification field should be omitted when nil.
	if contains(string(data), "classification") {
		t.Errorf("nil classification should be omitted: %s", data)
	}
}

func TestSideEffect_WithClassification(t *testing.T) {
	se := SideEffect{
		ID:          "se-a1b2c3d4",
		Type:        ReturnValue,
		Tier:        TierP0,
		Location:    "foo.go:10:1",
		Description: "returns int",
		Target:      "int",
		Classification: &Classification{
			Label:      Contractual,
			Confidence: 85,
			Signals: []Signal{
				{Source: "interface", Weight: 30},
			},
		},
	}

	data, err := json.Marshal(se)
	if err != nil {
		t.Fatalf("Marshal SideEffect: %v", err)
	}

	if !contains(string(data), "classification") {
		t.Errorf("classification should be present: %s", data)
	}

	var got SideEffect
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal SideEffect: %v", err)
	}

	if got.Classification == nil {
		t.Fatal("Classification should not be nil after unmarshal")
	}
	if got.Classification.Label != Contractual {
		t.Errorf("label = %q, want %q",
			got.Classification.Label, Contractual)
	}
}

func TestClassificationLabel_Values(t *testing.T) {
	if Contractual != "contractual" {
		t.Errorf("Contractual = %q, want %q", Contractual, "contractual")
	}
	if Incidental != "incidental" {
		t.Errorf("Incidental = %q, want %q", Incidental, "incidental")
	}
	if Ambiguous != "ambiguous" {
		t.Errorf("Ambiguous = %q, want %q", Ambiguous, "ambiguous")
	}
}

// contains is a simple helper for substring checks in tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFunctionTarget_QualifiedName(t *testing.T) {
	tests := []struct {
		name     string
		target   FunctionTarget
		expected string
	}{
		{
			name:     "package function",
			target:   FunctionTarget{Function: "ParseConfig"},
			expected: "ParseConfig",
		},
		{
			name:     "pointer receiver method",
			target:   FunctionTarget{Function: "Save", Receiver: "*Store"},
			expected: "(*Store).Save",
		},
		{
			name:     "value receiver method",
			target:   FunctionTarget{Function: "String", Receiver: "Config"},
			expected: "(Config).String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.target.QualifiedName()
			if got != tt.expected {
				t.Errorf("QualifiedName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestSideEffectType_NeutralAliases verifies that language-neutral
// aliases are equal to their Go-specific counterparts.
func TestSideEffectType_NeutralAliases(t *testing.T) {
	tests := []struct {
		neutral  SideEffectType
		goName   SideEffectType
		label    string
	}{
		{AsyncTaskSpawn, GoroutineSpawn, "AsyncTaskSpawn == GoroutineSpawn"},
		{AsyncMessageSend, ChannelSend, "AsyncMessageSend == ChannelSend"},
		{AsyncChannelClose, ChannelClose, "AsyncChannelClose == ChannelClose"},
		{BarrierOp, WaitGroupOp, "BarrierOp == WaitGroupOp"},
		{PanicRecovery, RecoverBehavior, "PanicRecovery == RecoverBehavior"},
		{FFICall, CgoCall, "FFICall == CgoCall"},
		{ObjectPoolOp, SyncPoolOp, "ObjectPoolOp == SyncPoolOp"},
	}

	for _, tt := range tests {
		if tt.neutral != tt.goName {
			t.Errorf("%s: %q != %q", tt.label, tt.neutral, tt.goName)
		}
	}
}
