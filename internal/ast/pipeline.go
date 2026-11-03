package ast

// PipelineClaimStmt represents:
// [@multi_device] pipeline_claim N stages micro_batch M into plan_var ·
type PipelineClaimStmt struct {
	Pos_        Position
	IsAnnotated bool
	Stages      int
	MicroBatch  int
	Into        string
}

func (s *PipelineClaimStmt) Pos() Position { return s.Pos_ }
func (s *PipelineClaimStmt) stmtNode()     {}

// PipelineRunStmt represents:
// pipeline_run plan_var on input into output ·
type PipelineRunStmt struct {
	Pos_   Position
	Plan   string
	Input  Expr
	Output Expr
}

func (s *PipelineRunStmt) Pos() Position { return s.Pos_ }
func (s *PipelineRunStmt) stmtNode()     {}

