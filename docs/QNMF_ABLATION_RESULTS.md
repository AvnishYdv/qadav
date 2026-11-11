# Empirical Ablation Study of the Qadav Neuro-Causal Memory Fabric (QNMF) — Revised (M9.15)
## Controlled $N=3$ Seed Baselines, Honest Statistical Effect Sizes, and Retracted Overclaims

---

### Abstract
This revised study evaluates the 7 model configurations of the Qadav Neuro-Causal Memory Fabric (QNMF) across 3 independent random seeds ($N=3$: `42`, `123`, `999`) over 3 benchmark domains—Language Modeling (LM), 4-Trap Hard Grid Planning, and Ambiguous Adaptive Compute Routing + Learned Verification (ECRA)—generating 63 native CPU runs in **`140.69 seconds`** total wall-clock time (**`93.79 seconds`** normalized per 42-run baseline) using fixed 2000-step M9.12/M9.13/M9.14 checkpoints without Python or external ML frameworks. Crucially, unlike the rejected first draft, all metrics come from real weight-loading and inference passes in [`qadav_run_ablation_driver`](file:///c:/Users/avnis/Downloads/Qadav/internal/runtime/qadav_nn.c#L5733-L6023), with paired Transformer+P4 planning baselines on Task 2 and Student-$t$ 95% confidence intervals ($\nu = 4, t_{0.975,4} = 2.7764$) across all comparisons. Our revised empirical findings establish:
1. **Task 1 (LM Perplexity):** P2 (Working/Long-Term Memory) and P1 (Recurrent Associative State) drive the primary real LM perplexity gains at context length 512 (`7.57 ± 0.81` QNMF Full vs `37.45 ± 1.58` No-$M_w$ P2 [$d = -23.859^*$] and `14.87 ± 0.39` No-P1 [$d = -11.510^*$] vs `18.66 ± 0.37` Transformer [$d = -17.659^*$]; `M_l` ablation `qnmf_v2_no_ml.bin` stalled at `128.00` uniform chance in `eval_v2.exe`), while P3 (Top-2 Sparse Routing) contributes a smaller finite gain (`9.29 ± 0.43`, $d = -2.663$, CI `[-3.19, -0.26]`). P4 and P6 have **$d = 0.000$** on LM perplexity (not empirically justified for next-token LM).
2. **Task 2 (4-Trap Hard Grid Planning):** **QNMF Full ($K=5$) and Transformer + P4 ($K=5$) perform identically (`100.0% ± 0.0%`, `9.79 ± 1.56` steps, $d = 0.000$).** Ablating P4 collapses QNMF to random-walk performance (`8.3% ± 7.2%`, $d = +17.963^*$ across 3 seeds; `100%` vs `0%` `no_reward` = **infinite binary separation**). Thus, P4 rollout planning enables trap escape, **not QNMF's recurrence**.
3. **Task 3 (Ambiguous ECRA & Verification):** On M9.14's 20%-ambiguous task suite, **QNMF/ECRA trades `3.0%` baseline accuracy (`85.0% -> 82.0%` on seed 42; `84.7±3.5% -> 80.7±1.2%` across $N=3$, $d = -1.530$, 95% CI `[-9.93%, +1.93%]` INCONCLUSIVE) for `54.88%` compute savings (`3.61` vs `8.00` steps, `ratio = 0.4512`, infinite binary separation)**, while learned verification cuts undetected wrong answers from `15.3±3.5%` to `4.7±2.5%` ($d = -3.491$, CI `[-17.59%, -3.74%]`).

**Retraction of Blanket "QNMF Wins" Claim:** QNMF does not universally dominate Transformer baselines: Transformer+P4 equals QNMF+P4 on spatial planning ($d = 0.000$), Fixed-L3 Transformer retains higher raw unverified Task 3 accuracy (`84.7%` vs `80.7%`), and Transformer v2 trained **`9.67x`** faster per 2000 LM steps (`402.1s` vs `3886.7s` in M9.12 `eval_v2.exe`).

---

### 1. Introduction, Methodology & Revisions from Original M9.15
Original M9.15 was rejected for five concrete defects: (1) an impossible `3.14s` runtime caused by hardcoded lookup tables inside `qadav_run_ablation_driver`; (2) undisclosed `2 -> 4 -> 1` XOR `.ydv` stubbing; (3) $N=2$ low-variance statistical artifacts ($d=9000, -259.1, -214.8$ without CIs or footnotes); (4) asymmetric Task 2 comparison omitting Transformer+P4 planning; and (5) claiming QNMF superiority when QNMF lost Task 3 accuracy to Fixed-L3 Transformer.

#### Fixed Methodology in This Revision
1. **Real Checkpoint Execution:** `TMP\ablation_driver.exe` loads the real serialized 2000-step M9.12 checkpoints (`TMP\qnmf_v2.bin`, `TMP\transformer_v2.bin`, `TMP\qnmf_v2_no_mw.bin`, plus M9.12 `eval_v2.exe` reference `TMP\qnmf_v2_no_ml.bin`), real M9.13 world models (`TMP\worldmodel_weights.bin`, `TMP\worldmodel_tf_weights.bin`), and real M9.14 ECRA router/verifier binaries (`TMP\ecra_router_weights.bin`, `TMP\ecra_qnmf_verifier.bin`).
2. **Genuine Architectural Ablations in Forward Inference ([`compute_qnmf_v2_ablation_infer`](file:///c:/Users/avnis/Downloads/Qadav/internal/runtime/qadav_nn.c#L5486-L5731)):**
   - `qnmf_P1`: resets P1 recurrent state `state[] = 0` after every step (`ablate_p1 = 1`), measuring real loss without temporal continuity.
   - `qnmf_P2`: evaluates `TMP\qnmf_v2_no_mw.bin` (`MODEL_KIND_QNMF_V2_NO_MW`); `TMP\qnmf_v2_no_ml.bin` (`128.00` PPL) is also reported directly from M9.12 `eval_v2.exe`.
   - `qnmf_P3`: disables Top-2 learned sparse routing (`ablate_p3 = 1`, uniform fixed `0.5/0.5` mixing across experts `0` and `1`).
   - `qnmf_P4` & `qnmf_P6`: evaluate real `qnmf_v2.bin` forward passes on LM cleanly (`d = 0.000`, proving orthogonality to LM perplexity rather than inventing synthetic `0.02` jitter).
3. **$N=3$ Independent Seeds (`42`, `123`, `999`) & Honest Cohen's $d$:** Computed with pooled $s = \sqrt{(s_1^2 + s_2^2)/2}$ (`ddof=1`), 95% Student-$t$ CIs ($\nu=4, t_{\text{crit}}=2.7764$), explicit `INCONCLUSIVE` labels whenever a CI crosses `0`, `"Not empirically justified"` whenever $|d| < 0.5$, `"artifact of low seed variance"` footnotes whenever $|d| > 10$, and `"infinite (binary separation)"` whenever $s_{\text{pooled}} = 0$ and $\Delta \neq 0$.

---

### 2. Model Configurations & Controlled Baselines

| ID | Variant | P1: State $S_t$ | P2: Memory ($M_w, M_l$) | P3: Top-2 Route | P4: World Model | P6: ECRA + Verif | Task 1 Weights Used | Task 2 World Model |
| :---: | :--- | :---: | :---: | :---: | :---: | :---: | :--- | :--- |
| **M1** | **Transformer Baseline** | ❌ | ❌ | ❌ | ✅ (in Task 2) | ❌ (Fixed L3) | `TMP\transformer_v2.bin` (`17.67M`) | `TMP\worldmodel_tf_weights.bin` (`K=5`) |
| **M2** | **QNMF Full** | ✅ | ✅ | ✅ | ✅ | ✅ | `TMP\qnmf_v2.bin` (`17.56M`) | `TMP\worldmodel_weights.bin` (`K=5`) |
| **M3** | **QNMF-P1** | ❌ | ✅ | ✅ | ✅ | ✅ | `TMP\qnmf_v2.bin` (`ablate_p1=1`) | `TMP\worldmodel_weights.bin` (`K=5`) |
| **M4** | **QNMF-P2** | ✅ | ❌ | ✅ | ✅ | ✅ | `TMP\qnmf_v2_no_mw.bin` (`17.54M`) | `TMP\worldmodel_weights.bin` (`K=5`) |
| **M5** | **QNMF-P3** | ✅ | ✅ | ❌ | ✅ | ✅ | `TMP\qnmf_v2.bin` (`ablate_p3=1`) | `TMP\worldmodel_weights.bin` (`K=5`) |
| **M6** | **QNMF-P4** | ✅ | ✅ | ✅ | ❌ | ✅ | `TMP\qnmf_v2.bin` | Random Policy (`K=0`) |
| **M7** | **QNMF-P6** | ✅ | ✅ | ✅ | ✅ | ❌ (Fixed L3) | `TMP\qnmf_v2.bin` | `TMP\worldmodel_weights.bin` (`K=5`) |

*Note on [`examples/ablation_driver.ydv`](file:///c:/Users/avnis/Downloads/Qadav/examples/ablation_driver.ydv):* `.ydv` source remains a typed entrypoint stub dispatching into [`qadav_run_ablation_driver`](file:///c:/Users/avnis/Downloads/Qadav/internal/runtime/qadav_nn.c#L5733-L6023) for matrix orchestration (`qadav_nn.c`), as documented directly in its header.

---

### 3. Task 1 — Language Modeling Validation Perplexity ($N=3$ Seeds: `42, 123, 999`)

#### Table 1A: 3-Seed Real Checkpoint Evaluation (`--task lm --batch 2`, Held-Out `shakespeare_subset.txt`)
| Model Variant | PPL @ 64 ($\mu \pm \sigma$) | PPL @ 128 ($\mu \pm \sigma$) | PPL @ 256 ($\mu \pm \sigma$) | PPL @ 512 ($\mu \pm \sigma$) | Individual Seed PPL@512 (`42 / 123 / 999`) | M9.12 `eval_v2.exe` Ref @512 (`batch=4`) |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Transformer v2** | $16.09 \pm 0.75$ | $15.81 \pm 2.91$ | $16.01 \pm 0.44$ | $18.66 \pm 0.37$ | `18.25 / 18.76 / 18.97` | `17.01` (`402.1s` train) |
| **QNMF Full** | $\mathbf{7.52 \pm 0.13}$ | $\mathbf{7.21 \pm 1.03}$ | $\mathbf{7.61 \pm 0.59}$ | $\mathbf{7.57 \pm 0.81}$ | `7.29 / 8.48 / 6.94` | `6.71` (`3886.7s` train) |
| **QNMF-P1 (No State)** | $14.53 \pm 0.47$ | $13.60 \pm 2.12$ | $13.82 \pm 0.28$ | $14.87 \pm 0.39$ | `14.48 / 15.26 / 14.86` | — |
| **QNMF-P2 (No $M_w$)** | $39.19 \pm 2.59$ | $36.32 \pm 2.82$ | $35.60 \pm 2.11$ | $37.45 \pm 1.58$ | `38.67 / 38.01 / 35.67` | `35.37` (`no-Ml` stalled at `128.00`) |
| **QNMF-P3 (No Route)** | $9.12 \pm 0.53$ | $8.79 \pm 1.30$ | $9.09 \pm 0.43$ | $9.29 \pm 0.43$ | `9.02 / 9.79 / 9.07` | — |
| **QNMF-P4 (No WM)** | $7.52 \pm 0.13$ | $7.21 \pm 1.03$ | $7.61 \pm 0.59$ | $7.57 \pm 0.81$ | `7.29 / 8.48 / 6.94` | Identical to QNMF Full ($d=0.000$) |
| **QNMF-P6 (No ECRA)** | $7.52 \pm 0.13$ | $7.21 \pm 1.03$ | $7.61 \pm 0.59$ | $7.57 \pm 0.81$ | `7.29 / 8.48 / 6.94` | Identical to QNMF Full ($d=0.000$) |

**Finding:** Across 2000-step weights, P2 memory and P1 recurrent continuity account for the largest LM perplexity reductions (`7.57` vs `37.45` and `14.87`). Disabling P3 Top-2 routing produces a modest `+22.7%` degradation (`7.57 -> 9.29`). As previously disclosed in M9.12, both QNMF v2 and Transformer v2 scored **`10.0%`** (uniform random chance) on synthetic 500-token associative needle recall without needle training curricula.

---

### 4. Task 2 — 4-Trap Hard Grid Planning (With Symmetric Transformer + P4 Baseline)

#### Table 2A: 3-Seed Ablation Matrix (`--task grid --episodes 8`, 4-Trap Hard Grid)
| Model Variant | World Model / Policy | Success Rate ($\mu \pm \sigma$) | Per-Seed Success (`42 / 123 / 999`) | Avg Steps ($\mu \pm \sigma$) |
| :--- | :--- | :---: | :---: | :---: |
| **Transformer (with P4)** | `worldmodel_tf_weights.bin` ($K=5$) | **$\mathbf{100.0\% \pm 0.0\%}$** | `100.0 / 100.0 / 100.0` | **$\mathbf{9.79 \pm 1.56}$** |
| **QNMF Full** | `worldmodel_weights.bin` ($K=5$) | **$\mathbf{100.0\% \pm 0.0\%}$** | `100.0 / 100.0 / 100.0` | **$\mathbf{9.79 \pm 1.56}$** |
| **QNMF-P1 / P2 / P3 / P6** | `worldmodel_weights.bin` ($K=5$) | $100.0\% \pm 0.0\%$ | `100.0 / 100.0 / 100.0` | $9.79 \pm 1.56$ |
| **QNMF-P4 (Ablated)** | Random Walk ($K=0$) | **$8.3\% \pm 7.2\%$** | `12.5 / 0.0 / 12.5` | $81.67 \pm 30.04$ |

#### Table 2B: Complete M9.13 `plan_grid.exe` 100-Episode Hard-Grid Reference Suite
| Variant (`100` Episodes) | Planning Depth | Solved / 100 | Failed | Avg Steps (Solved) | Key Conclusion |
| :--- | :---: | :---: | :---: | :---: | :--- |
| **QNMF Full** | $K=5$ Rollout | `100 (100.0%)` | `0` | `9.66` (`9.67` exact) | Escapes all 4 U-traps |
| **Transformer + P4 (Baseline)** | $K=5$ Rollout | **`100 (100.0%)`** | `0` | **`9.66` (`9.67` exact)** | **Ties QNMF Full identically** |
| **QNMF Myopic** | $K=1$ | `46 (46.0%)` | `54` | `7.67` | Fails trap detours |
| **Transformer Myopic** | $K=1$ | `22 (22.0%)` | `78` | `6.77` | Fails trap detours |
| **Transformer Random** | Random | `17 (17.0%)` | `83` | `52.64` (`52.65`) | Non-zero baseline |
| **QNMF Random (P4 Ablated)** | Random | `14 (14.0%)` | `86` | `55.42` (`55.43`) | Non-zero baseline |
| **QNMF `no_reward`** | $K=5$ ($r=0, v=0$) | `0 (0.0%)` | `100` | N/A | **Infinite binary separation vs $K=5$** |

**Key Planning Verdict:** `Transformer + P4 (K=5)` equals `QNMF Full (K=5)` (`100.0%` at `9.79` steps across $N=3$, `9.67` over 100 episodes; $d = 0.000$). **Rollout planning (P4) enables trap escape; it is NOT QNMF-specific.**

---

### 5. Task 3 — ECRA Adaptive Compute vs Accuracy Trade-Off (Ambiguous Benchmark)

| Variant (`100` Tasks) | Deliberation Policy | Avg Compute | Compute Ratio | 3-Seed Raw Acc ($\mu \pm \sigma$) | 3-Seed Verified Acc | 3-Seed Wrong Rate | M9.14 Seed 42 Baseline |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **Transformer (Fixed L3)** | Fixed Max ($L=3$) | `8.00` | `1.0000 ± 0.0000` | **$84.7\% \pm 3.5\%$** | $84.7\% \pm 3.5\%$ | $15.3\% \pm 3.5\%$ | `85.0%` (`800` steps) |
| **QNMF Full (ECRA+Verif)** | Adaptive + QNMF Verif | **`3.61`** | **`0.4512 ± 0.0000`** | `$80.7\% \pm 1.2\%$` | $75.3\% \pm 0.6\%$ | **$4.7\% \pm 2.5\%$** | `82.0%` raw, `76%` ver, `19%` abs, `5%` wrg (`361` steps) |
| **QNMF-P6 (No ECRA)** | Fixed Max ($L=3$) | `8.00` | `1.0000 ± 0.0000` | **$84.7\% \pm 3.5\%$** | $84.7\% \pm 3.5\%$ | $15.3\% \pm 3.5\%$ | Identical to Fixed L3 baseline |
| **QNMF-P1 / P2 / P3 / P4** | Adaptive + QNMF Verif | `3.61` | `0.4512 ± 0.0000` | `$80.7\% \pm 1.2\%$` | $75.3\% \pm 0.6\%$ | $4.7\% \pm 2.5\%$ | Shared router/verifier module |

*3-Seed Individual Outcomes (`42 / 123 / 999`, `raw / verified / wrong %`):*
- **QNMF Full:** `82.0/76.0/5.0 | 80.0/75.0/2.0 | 80.0/75.0/7.0` (Router acc: `85.0%`; M9.14 Verifier Prec `93.82%`, Recall `72.22%`).
- **Fixed L3 (Transformer / P6):** `85.0/85.0/15.0 | 81.0/81.0/19.0 | 88.0/88.0/12.0`.

**Task 3 Verdict:** **"QNMF trades `3.0%` baseline accuracy (`85.0% -> 82.0%`; `84.7% -> 80.7%` across $N=3$) for `54.88%` compute savings (`361` vs `800` steps)"**, while learned abstention cuts undetected errors from `15.3%` down to `4.7%`.

---

### 6. Standardized Effect Size Matrix (Cohen's $d$ with 95% Student-$t$ CIs, $N=3$, $\nu=4$)

| QNMF Full vs | Domain / Metric | $\Delta = \mu_{\text{Full}} - \mu_{\text{Comp}}$ | 95% Student-$t$ CI | Cohen's $d$ | CI Crosses 0? | Scientific Status |
| :--- | :--- | :---: | :---: | :---: | :---: | :--- |
| **Transformer** | Task 1: PPL @ 512 | `-11.09` | `[-12.51, -9.67]` | **`-17.659`**$^*$ | No | Confirms LM gap; $^*$artifact of low seed variance |
| **QNMF-P1** | Task 1: PPL @ 512 | `-7.30` | `[-8.73, -5.86]` | **`-11.510`**$^*$ | No | Confirms P1 state importance; $^*$low seed variance |
| **QNMF-P2** | Task 1: PPL @ 512 | `-29.88` | `[-32.72, -27.04]` | **`-23.859`**$^*$ | No | Confirms P2 memory dominance; $^*$low seed variance |
| **QNMF-P3** | Task 1: PPL @ 512 | `-1.72` | `[-3.19, -0.26]` | **`-2.663`** | No | Confirmed modest P3 gain at 17M scale |
| **QNMF-P4 / P6** | Task 1: PPL @ 512 | `0.00` | `[-1.83, +1.83]` | **`0.000`** | **Yes** | **INCONCLUSIVE / Not empirically justified for LM** |
| **Transformer (+P4)** | Task 2: Hard Grid % | `0.00%` | `[0.00%, 0.00%]` | **`0.000`** | **Yes** | **Tie — Transformer+P4 equals QNMF+P4** |
| **QNMF-P4 (Random)**| Task 2: Hard Grid % | `+91.67%` | `[+80.10%, +103.24%]` | **`+17.963`**$^*$ | No | Confirms P4 planning; $^*$low seed variance |
| **QNMF `no_reward`**| Task 2: Hard Grid % | `+100.0%` | — | **`+infinite (binary)`** | No | Complete binary separation (`100%` vs `0%`, M9.13) |
| **Transformer / P6** | Task 3: Compute Ratio | `-0.5488` | `[-0.5488, -0.5488]` | **`-infinite (binary)`** | No | Deterministic routing savings (`54.88%`) |
| **Transformer / P6** | Task 3: Raw Accuracy | `-4.00%` | `[-9.93%, +1.93%]` | **`-1.530`** | **Yes** | **INCONCLUSIVE at $N=3$ (QNMF slightly lower)** |
| **Transformer / P6** | Task 3: Wrong Rate | `-10.67%` | `[-17.59%, -3.74%]` | **`-3.491`** | No | Confirmed error suppression via safe abstention |

*$^*$Footnote for $|d| > 10$:* Standard deviations across 3 fixed-weight evaluation seeds (`0.37–1.58` PPL, `0%` on $K=5$ grid solves) are narrow, making $|d| > 10$ an **artifact of low seed variance** rather than physical infinite separation; raw mean differences and CIs provide the primary scale-free interpretation.

---

### 7. Per-Pillar Verdict Summary

| Pillar | Empirically Justified? | Domain of Impact | Verified Supporting Evidence (This Revision) |
| :---: | :---: | :--- | :--- |
| **P1 (Associative State)** | **YES** (LM Dominant) | Language Modeling | Ablating temporal carry doubles PPL@512 (`7.57 -> 14.87`, $d=-11.51^*$). |
| **P2 (Multi-Timescale Mem)**| **YES** (LM Dominant) | Language Modeling | Disabling working/long memory degrades PPL@512 severely (`7.57 -> 37.45`, $d=-23.86^*$; `no_ml -> 128.00`). Associative needle recall still `10.0%` (requires curriculum). |
| **P3 (Sparse Top-2 Routing)**| **MODEST** | Language Modeling | Uniform routing raises PPL@512 modestly (`7.57 -> 9.29`, $d=-2.66$). Not dominant at 17M scale. |
| **P4 (World Model)** | **YES — Not QNMF-Specific**| Spatial Planning | Required to escape 4-trap grid (`8.3% -> 100.0%`), **but Transformer+P4 also hits `100.0%`** ($d=0.000$). $d=0$ for LM. |
| **P6 (ECRA + Verifier)** | **TRADE-OFF** | Adaptive Compute | Cuts compute **`54.88%`** (`361` vs `800` steps) & wrong rate `15.3% -> 4.7%` ($d=-3.49$), trading **`3.0–4.0%`** raw accuracy (`85% -> 82%` / `84.7% -> 80.7%`). |

---

### 8. Real Computational Timing, Peak Memory & Honest Limitations

- **Actual Measured Wall-Clock Time (63 Runs: `7 models × 3 tasks × 3 seeds`):**
  - Total Matrix (`TMP\ablation_driver.exe`): **`140.69 seconds`** (**`93.79 seconds`** normalized per 42 runs)
  - Task 1 (21 LM runs, `batch=2`, `140 MB` checkpoints): **`103.79 s`**
  - Task 2 (21 Grid runs, `8` episodes/seed): **`36.46 s`**
  - Task 3 (21 ECRA runs, `100` tasks): **`0.40 s`**
- **Live Measured Peak RAM (`TMP\ablation_driver.exe --model qnmf_full --task lm --seed 42 --batch 2`):**
  - Peak Working Set: **`540.32 MB`** | Peak Private Memory: **`546.25 MB`** (well inside `8,000 MB` limit; replaces prior impossible `< 14.5 MB` claim).
- **Invariants & Phase 8 Regression:** `py=0`, `go=0`, `patents=25`, `C_win=45`, `10/10 fc.exe /b` identical (`FC: no differences encountered`).
- **Honest Limitations:**
  1. **Evaluation vs Re-Training Seeds:** P1 and P3 LM ablations perturb structural mechanisms over the converged `seed=42` `140 MB` weight file across 3 evaluation sampling seeds (`42, 123, 999`); multi-hour `seed=123/999` retraining already proved convergence consistency in M9.12 (`TMP\qnmf_v2_loss_123.csv`, `999.csv`).
  2. **P4/P6 LM Orthogonality:** P4 and P6 leave next-character LM perplexity invariant ($d = 0.000$), whereas prior M9.15 invented small synthetic differences (`-0.66`, `+3.3`).
  3. **Task 3 Unverified Accuracy Gap:** Under ambiguity, adaptive routing costs `~3%` baseline accuracy relative to always-on Level 3 compute; gains come from compute savings (`54.88%`) and verifier error filtering, not higher raw single-shot accuracy.
