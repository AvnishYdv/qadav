package ast

// AADCertStmt represents generating an attested decoder training certificate:
// aad_cert with model ↦ model_bytes train ↦ train_bytes max_syndrome ↦ 64 max_corr_dist ↦ 8 into cert ·
type AADCertStmt struct {
	Pos_        Position
	Model       Expr
	TrainSet    Expr
	MaxSyndrome Expr
	MaxCorrDist Expr
	Into        string
}

func (s *AADCertStmt) Pos() Position { return s.Pos_ }
func (s *AADCertStmt) stmtNode()     {}

// AADAttestStmt represents attesting an operating decode envelope to a host QERA chain:
// aad_attest cert cert ↦ cert decode ↦ decoded into envelope ·
type AADAttestStmt struct {
	Pos_   Position
	Cert   string
	Decode string
	Into   string
}

func (s *AADAttestStmt) Pos() Position { return s.Pos_ }
func (s *AADAttestStmt) stmtNode()     {}

// MetaFitStmt represents binding a base certificate to a meta-learning state:
// meta_fit into meta_state cert ↦ cert ·
type MetaFitStmt struct {
	Pos_ Position
	Into string
	Cert string
}

func (s *MetaFitStmt) Pos() Position { return s.Pos_ }
func (s *MetaFitStmt) stmtNode()     {}

// MetaAdaptStmt represents adapting meta-learning state with an attested envelope:
// meta_adapt meta_state with envelope into adapted ·
type MetaAdaptStmt struct {
	Pos_     Position
	State    string
	Envelope string
	Into     string
}

func (s *MetaAdaptStmt) Pos() Position { return s.Pos_ }
func (s *MetaAdaptStmt) stmtNode()     {}

// MetaVerifyStmt represents verifying the cumulative adaptation chain of a meta-learning state:
// meta_verify meta_state into is_valid ·
type MetaVerifyStmt struct {
	Pos_  Position
	State string
	Into  string
}

func (s *MetaVerifyStmt) Pos() Position { return s.Pos_ }
func (s *MetaVerifyStmt) stmtNode()     {}

// CorruptAADEnvStmt represents corrupting an AAD envelope for tamper testing (--test only):
// corrupt_aad_env envelope at ↦ 16 ·
type CorruptAADEnvStmt struct {
	Pos_   Position
	Env    string
	Offset Expr
}

func (s *CorruptAADEnvStmt) Pos() Position { return s.Pos_ }
func (s *CorruptAADEnvStmt) stmtNode()     {}
