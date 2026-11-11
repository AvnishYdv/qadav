# Qadav Neuro-Causal Memory Fabric (QNMF): Architecture, Native Compilation, Controlled CPU Ablations, and Empirical Trade-Offs

**Project:** Qadav (`.ydv`) — Independent AI/AGI-Native Systems Programming Language  
**Phase / Milestone:** Phase 9 Capstone — M9.16 Revision (`M9.9–M9.15` Synthesis, Honest Reframing)  
**Target Environment:** Windows 10/11 x86_64 Native Binary Runtime (Zero-Libc, Single-Host CPU)  
**Compiler Foundation:** Self-Hosted `bin/qyn.exe` (Phase 8 4/4 Fixed-Point Bootstrap Verified)  
**Date:** November 2026  

---

## Abstract

We present the formal architectural specification, freestanding compiler integration, and controlled empirical evaluation of the **Qadav Neuro-Causal Memory Fabric (QNMF)** inside the Qadav (`.ydv`) zero-libc native systems programming language. QNMF was originally proposed across Milestones M9.9–M9.10 to address structural weaknesses in standard autoregressive Transformers and first-generation state-space models (SSMs): $O(L^2)$ attention KV-cache scaling, uniform $TC^0$ circuit depth bounds, asymmetric sequence reversal errors, monolithic weight overwriting during test-time adaptation, load collapse under repetitive Mixture-of-Experts (MoE) prompts, unverified token-space planning rollouts, and fixed-depth compute waste. QNMF organizes computation into six interacting components: **P1** commutative bilateral causal state dynamics, **P2** multi-timescale hierarchical memory ($M_s \to M_w \to M_l$), **P3** entropic Optimal Transport (Sinkhorn) cognitive routing, **P4** predictive latent world modeling, **P5** cryptographic proof-carrying verification (`QERA`/`TTPCR`/`QEEH`), and **P6** Evidence-Calibrated Compute Routing & Allocation (ECRA).

To guard against premature architectural claims, this M9.16 revision synthesizes strictly corrected results after iterative audit rejections and rebuilds across upstream milestones (`M9.1, M9.4, M9.7, M9.11, M9.12, M9.15`, alongside compiler/protocol hardening in `M9.4a, M9.9, M9.13, M9.14`). All neural experiments execute end-to-end on consumer Windows CPU hardware within **`540.32 MB`** peak live RSS (`546.25 MB` private commit) and zero external runtime libraries (`0` Python, `0` Go, `45` freestanding Windows C modules). Across `N=3` fixed-checkpoint random seeds (`42, 123, 999`; `63` evaluated runs completed in `140.69 s` total wall-clock time) and supporting 2,000-step single-CPU pretraining benchmarks (`0.31M–0.42M` toy and `17.56M–17.67M` scaled variants), four central empirical findings emerge:

1. **Sequence Modeling Convergence vs. Wall-Clock Throughput:** Under 2,000-step CPU pretraining on byte/character-level Shakespeare (`17.56M` QNMF v2 vs. `17.67M` Transformer v2), QNMF v2 reaches lower held-out perplexity at context `512` (**`6.71`** vs. **`17.01`** in M9.12; **`7.57 ± 0.81`** vs. **`18.66 ± 0.37`** across `N=3` M9.15 seeds), with Pillar 2 memory components driving the primary loss sensitivity (`no-Ml` stalls at `4.8520` / PPL `128.00`; `no-Mw` degrades to PPL `35.37` in M9.12 and `37.45 ± 1.58` in M9.15). However, scalar CPU execution of QNMF's sequential state recurrence, Sinkhorn routing, and memory slot updates is **`9.67x` slower per step** (`3,886.7 s` vs. `402.1 s` for 2,000 steps). Furthermore, P3 Sinkhorn routing produces only a modest perplexity benefit (`9.29 ± 0.43` without P3 vs. `7.57 ± 0.81` Full, Cohen's `d = -2.663`), and long-range **associative recall remains at `10.0%` random digit chance (`5/50` @ 500 tokens)** for all four `17.6M` models.
2. **Backbone-Agnostic World-Model Planning (`Transformer+P4 = QNMF+P4`, `d = 0.000`):** On the M9.13 Hard 4-Trap `8x8` grid (`18` walls), 100-episode rollout evaluation shows both **`QNMF+P4 (K=5)`** and **`Transformer+P4 (K=5)`** achieve **`100.0%` success** (`9.66`/`9.67` mean steps in M9.13; **`100.0 ± 0.0%`** at **`9.79 ± 1.56` steps** across `N=3` seeds in M9.15), yielding an exact **`d = 0.000`** difference (95% CI `[0.00, 0.00]`). Myopic `K=1` collapses to `46.0%` (QNMF) and `22.0%` (Transformer), while random / zero-reward policies fall to `8.3 ± 7.2%` / `0.0%`. Thus, P4 lookahead planning works effectively, **but confers no QNMF-exclusive planning advantage over a P4-equipped Transformer baseline**.
3. **ECRA Compute-Accuracy Trade-Off (`54.88%` Savings for `3%–4%` Raw Accuracy Cost):** Under M9.14 ambiguous task routing (`100` tasks, `20%` borderline ambiguity, `4%` label noise), the learned ECRA router (`129 -> 64 -> 4`) plateaus at a realistic `0.1265` cross-entropy floor (`85.0%` tier accuracy) and executes **`361 vs. 800` compute steps (`54.88%` compute savings; ratio `0.4512`)**. This efficiency trades off **`3.0%` raw accuracy** in M9.14 (`82.0%` vs. `85.0%` Fixed L3) and **`-4.00%` raw mean difference** in M9.15 `N=3` runs (`80.7 ± 1.2%` vs. `84.7 ± 3.5%`, `d = -1.530`, 95% CI `[-9.93%, +1.93%]` — statistically **`INCONCLUSIVE`**). Learned uncertainty verification lowers error rate from `15.3 ± 3.5%` to `4.7 ± 2.5%` (`d = -3.491`, 95% CI `[-17.59%, -3.74%]`), while tying equally (`93.82%` precision, `72.22%` recall) on both QNMF (`tau = 0.66`) and Transformer (`tau = 0.50`) traces.
4. **Native Cryptographic Systems Guarantees:** Unlike statistical ML trade-offs, Phase 9 cryptographic invariants (`M9.1–M9.8`, `M9.4` Frobenius norm fix `200.000000`, `M9.4a` compiler 2D tensor literal emission, `10/10` Phase 8 bitwise compiler identity) enforce deterministic fail-stop integrity across `612–614 KB` zero-libc Windows executables.

Taken together, the Phase 9 evidence establishes QNMF as a **proposed, partially validated research architecture with measurable toy- and single-CPU prototype behaviors and explicit trade-offs**, not a proven large-scale successor to frontier Transformers.

---

## 1. Introduction & Scientific Motivation

Contemporary machine learning research increasingly divides language modeling innovations from systems programming guarantees. Foundation model research generally assumes permissive Python runtimes (`libtorch`, CUDA shared objects, BLAS backends), opaque tensor memory allocations, post-hoc guardrail filters, and multi-terabyte GPU clusters before basic architectural invariants can even be compiled into standalone execution artifacts. Conversely, conventional systems languages (C, Rust, C++) deliver memory control and deterministic deployment without native primitives for uncertainty propagation (`confident<T>`), multi-timescale tensor state, cryptographically attested reasoning graphs (`QERA`), or compile-time/runtime proof witnesses (`TTPCR`).

Qadav (`.ydv`) is an experimental, local-first systems and AI programming language designed to close that divide inside an independent self-hosted compiler (`bin/qyn.exe`, bootstrapped identically through Stage 0 Go -> Stage 1 Qadav -> Stage 2 Self -> Stage 3 Fixed-Point in Phase 8) backed by **`45` zero-libc Windows C runtime modules** (linking only `kernel32.lib` via `clang` and `lld`) and **`25` documented system inventions** (`6` defensive public-domain publications and `19` proprietary patent-pending mechanisms in [`docs/PATENTS.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/PATENTS.md)).

### 1.1 The Phase 9 Scientific Integrity Mandate

During Phase 9, engineering proceeded in two linked tracks:
* **Track I (`M9.1–M9.8` + `M9.4a`): Systems & Cryptographic Invariants.** Packaging 8 AI security/safety inventions (`QEMWV`, `QERA`, `MCTP`, `TTPCR`, `BARAT`, `CDAGA`, `QECMA`, `QESI`) into verifiable `.ydv` native products plus an end-to-end 9-step integrated workflow (`product_integrated.exe`, `87 ms`), culminating in compiler bug resolution (`M9.4a`) for native 2D tensor literal lowering.
* **Track II (`M9.9–M9.16`): QNMF Scientific Hypothesis, Prototyping & Controlled Ablation.** Auditing prior literature (`M9.9`), formalizing the QNMF mathematical equations (`M9.10`), building and training `0.31M` toy models (`M9.11`) and `17.56M` scaled CPU models (`M9.12`), evaluating hard multi-trap world-model planning (`M9.13`), testing ambiguous adaptive compute routing (`M9.14`), and running a controlled `N=3` seed ablation matrix (`M9.15`).

Crucially, strict milestone review caught critical methodological flaws across initial submissions of Phase 9: omitted Phase 8 `fc /b` compiler checks (`M9.1`), an erroneous C runtime Frobenius norm calculation in Byzantine aggregation (`M9.4`, corrected from `0.06` to exact `sqrt(4 * 100^2) = 200.0`), silent compiler workaround on 2D matrices (`M9.4a`), noisy 20-step undertrained runs (`M9.11`, `M9.12`), trivially solvable open grids where lookahead depth $K=5$ could not separate from $K=1$ (`M9.13`), unrealistically separable `0.0000`-loss routing tasks (`M9.14`), and an initial `M9.15` report built on asymmetric baselines, `N=2` seed variance artifacts, and understated execution times. **Every rejected milestone was rebuilt from source without falsifying data or disabling checks.** This capstone report synthesizes exclusively the verified, post-revision evidence across `M9.1–M9.15`.

### 1.2 Summary of Key Verified Trade-Offs

Rather than claiming blanket superiority, the Phase 9 evidence exposes concrete architectural trade-offs summarized below:

* **Perplexity Convergence vs. Scalar CPU Speed (`M9.12`, `M9.15`):** On byte/character Shakespeare windows up to `512` tokens after `2,000` Adam steps, QNMF v2 (`17.56M`) achieves lower held-out cross-entropy loss/perplexity (`6.71` vs. `17.01` in M9.12; `7.57 ± 0.81` vs. `18.66 ± 0.37` in M9.15 `N=3`) than a parameter-matched Transformer v2 (`17.67M`), with `M_w` and `M_l` memory modules proving load-bearing (`no-Mw` degrades to PPL `35.37–37.45`; `no-Ml` does not move from uniform chance `128.00`). Yet QNMF v2 requires **`9.67x` longer wall-clock training time** on CPU (`3,886.7 s` vs. `402.1 s`) and fails at long-range **associative needle recall (`10.0%` = random guessing)**.
* **World-Model Planning Power Comes from P4, Not QNMF (`M9.13`, `M9.15`):** Adding a `92,418`-parameter latent transition-and-value world model (`P4`) with depth `K=5` rollout eliminates failure on a `4`-trap obstacle maze (`100.0%` vs. `22.0%–46.0%` at `K=1`), **regardless of whether the encoder backbone is QNMF or Transformer (`d = 0.000`)**.
* **Dynamic Routing Saves Compute at a Measured Accuracy Cost (`M9.14`, `M9.15`):** ECRA reduces compute allocation by **`54.88%`** (`0.4512` compute ratio) across mixed-difficulty prompts, at a measurable **`3.0%` accuracy trade-off** in M9.14 evaluation (`-4.00%` mean delta across `N=3` M9.15 seeds, 95% CI spanning zero, `INCONCLUSIVE`). High-risk hallucinations are primarily suppressed by post-hoc learned verification + abstention (`19%–23%`), which works equivalently (`93.82%` precision) on QNMF and Transformer outputs.

---

## 2. Prior-Art Audit & Verification Resolution (M9.9)

Milestone `M9.9` (`docs/QNMF_PRIOR_ART_AUDIT.md`) established empirical constraints in six prior-art domains to ensure QNMF did not reinvent existing mechanisms without documented justification. During revision auditing, two rigorous bibliographic cleanups were enforced:

1. **Purge of Unverified Citations:** Three unverified entries present in earlier pre-revision drafts were confirmed missing/unresolvable during audit checks and are **permanently excluded** from this report and bibliography.
2. **Live HTTP 200 & Metadata Verification of Five 2026-Dated arXiv Citations:** Because five core M9.9 references carry 2025/2026 arXiv identifiers (`2610.19709`, `2603.15994`, `2602.14159`, `2512.23995`, `2609.04453`), we executed a live HTTP verification probe against their arXiv abstract endpoints (`November 11, 2026`) capturing exact status codes, HTML `<title>` tags, and `<meta name="citation_author">` fields. Table 1 presents the verbatim HTML verification record alongside a critical author-attribution correction.

### Table 1: Live HTTP Verification of Recent arXiv References (Executed November 11, 2026)

| arXiv ID | HTTP Status | Verbatim arXiv `<title>` Tag | Verified `<meta citation_author>` & Date | Audit Resolution & Canonical Anchoring |
|---|---|---|---|---|
| `arXiv:2610.19709` | **200 OK** | `[2610.19709] Beyond State Capacity: Why Retrieval Breaks in Mamba` | Yarden, Nir; Ben-Artzy, Amit; Schwartz, Roy (`2026/10/14`) | **VERIFIED.** Grounded in canonical retrieval/SSM bounds (Gu & Dao 2023; Liu et al. 2023). |
| `arXiv:2603.15994` | **200 OK** | `[2603.15994] Selective Memory for Artificial Intelligence: Write-Time Gating with Hierarchical Archiving` | **Zahn, Oliver; Chana, Simran** (`2026/03/16`) | **VERIFIED URL / AUTHOR CORRECTED.** Prior draft misattributed `"Liu et al."`; live arXiv metadata is **Zahn & Chana (2026)**. Grounded additionally in Behrouz et al. (2025, Titans) & Dohare et al. (2023/2024). |
| `arXiv:2602.14159` | **200 OK** | `[2602.14159] Synergistic Intra- and Cross-Layer Regularization Losses for MoE Expert Specialization` | Hu, Rizhen; Cao, Yuan; Kong, Boao; Sun, Mou; Yuan, Kun (`2026/02/15`) | **VERIFIED.** Grounded in canonical sparse MoE routing (Shazeer et al. 2017; Fedus et al. 2022; DeepSeek-AI 2024). |
| `arXiv:2512.23995` | **200 OK** | `[2512.23995] RepetitionCurse: Measuring and Understanding Router Imbalance in Mixture-of-Experts LLMs under DoS Stress` | Huang, Ruixuan; Wang, Qingyue; Huang, Hantao; Gao, Yudong; Chen, Dong; Wang, Shuai; Wang, Wei (`2025/12/30`) | **VERIFIED.** Grounded in MoE capacity bottlenecks (Fedus et al. 2022; Shazeer et al. 2017). |
| `arXiv:2609.04453` | **200 OK** | `[2609.04453] When Load-Balancing Goes Too Far: Expert Pruning in Over-Dispersed Mixture-of-Experts Models` | Kapusuzoglu, Berkcan; Pryor, Connor; Cho, Sangwoo; Chakraborty, Supriyo; Zhang, Shi-Xiong; Sahu, Sambit; Naphade, Milind (`2026/09/03`) | **VERIFIED.** Grounded in sparse MoE trade-offs (Fedus et al. 2022; DeepSeek-AI 2024). |

### 2.1 Mapping Foundational Gaps to Canonical Literature & Phase 9 Results

Every theoretical mechanism in QNMF rests directly on peer-reviewed / canonical foundations rather than relying solely on recent 2026 preprints:

1. **Quadratic Context Memory vs. Fixed State Compression (`Vaswani et al., 2017`; `Dao, 2023`; `Gu & Dao, 2023`; `Dao & Gu, 2024`; `Liu et al., 2023`):** Standard multi-head attention incurs $\Omega(L^2 d)$ compute and $\Omega(L)$ KV-cache memory. Selective state-space models (Mamba/Mamba-2) recover $O(L)$ inference scaling, yet suffer finite-state retrieval degradation ("Lost in the Middle", `Liu et al., 2023`; `Yarden et al., 2026`). *Phase 9 outcome:* At `512` context on Shakespeare, QNMF converges cleanly, **but needle-in-a-haystack associative recall at `500` tokens remains at `10.0%` (`5/50` chance)** across both QNMF v2 and Transformer v2.
2. **Circuit Complexity & Reversal / Asymmetric Biases (`Merrill & Sabharwal, 2023`; `Berglund et al., 2024`; `Chen et al., 2025`; `Marcus, 2020`):** Fixed-depth log-precision Transformers reside inside uniform circuit class $TC^0$ (`Merrill & Sabharwal, 2023`), restricting exact state tracking without scratchpad Chains of Thought, while causal autoregression induces reversal asymmetry ($P(B|A) \not\Rightarrow P(A|B)$, `Berglund et al., 2024`) and causal 1D convolutions create directional sequence pathologies in Mamba (`Chen et al., 2025`).
3. **Test-Time Memory, Unlearning & Loss of Plasticity (`Behrouz et al., 2025`; `Dohare et al., 2023/2024`; `Zahn & Chana, 2026`):** Continuous gradient-updated test-time memory (Titans) conflates relational updates with parameter weights, risking loss of plasticity (`Dohare et al., 2023/2024`) and interference during targeted erasure (`Zahn & Chana, 2026`). QNMF separates working plasticity ($M_w$) from slotted storage ($M_l$). *Phase 9 outcome:* Ablating `M_w` and `M_l` severely injures language modeling convergence (`M9.12`/`M9.15`), though $O(1)$ multi-task erasure protocols (`EXP-3`, `EXP-5`) were not executed in Phase 9.
4. **Mixture-of-Experts Routing Stability & Over-Regularization (`Shazeer et al., 2017`; `Fedus et al., 2022`; `DeepSeek-AI, 2024`; `Hu et al., 2026`; `Huang et al., 2025`; `Kapusuzoglu et al., 2026`):** Auxiliary balancing losses in softmax gating can fail under adversarial low-entropy token repetition (`Huang et al., 2025`) or hurt specialization (`Hu et al., 2026`; `Kapusuzoglu et al., 2026`). Optimal Transport / Sinkhorn balancing stabilizes expert dispatch theoretically. *Phase 9 outcome:* P3 routing lowers perplexity modestly (`9.29 -> 7.57`, `d=-2.663`), **while large-scale FLOP speedups and DoS stress benchmarks remain deferred to GPU cluster testing**.
5. **Latent World Models vs. Autoregressive Planning (`Ha & Schmidhuber, 2018`; `Hafner et al., 2023`; `Bardes et al., 2021`):** Planning in compact continuous latent states $z_t$ via VICReg-regularized forward dynamics (`Bardes et al., 2021`) avoids token-space drift. *Phase 9 outcome:* P4 depth-$K=5$ rollouts achieve **`100.0%`** hard-grid success—equally for QNMF and Transformer backbones (`d = 0.000`).
6. **Adaptive Computation (`Graves, 2016`; `Tay et al., 2022`; `Fu et al., 2024`):** Dynamic compute allocation scales test-time compute (`Graves, 2016`) selectively per input complexity. *Phase 9 outcome:* M9.14 ECRA trades `-3%` accuracy for `54.88%` step savings, proving real accuracy/compute Pareto curvature rather than free improvements.

### Table 2: Summary Audit of Phase 9 Prior-Art Claims vs. Observed Reality

| Gap Domain | Canonical Citations | QNMF Target Pillar | Phase 9 Empirical Status |
|---|---|---|---|
| Context Compute / Memory | `Vaswani 2017`, `Dao 2023`, `Gu & Dao 2023` | **P1 + P2** ($O(L)$ scan, fixed slots) | **PARTIAL:** Clean `512`-ctx convergence; `10%` associative recall (deferred) & `9.67x` slower CPU training |
| Plasticity & Memory Separation | `Behrouz 2025`, `Dohare 2024`, `Zahn & Chana 2026` | **P2** ($M_s \to M_w \to M_l$) | **LM SUPPORTED / UNLEARNING UNTESTED:** `no-Mw` + `no-Ml` strongly harm LM; `EXP-3`/`EXP-5` deferred |
| MoE Routing Stability | `Shazeer 2017`, `Fedus 2022`, `Huang 2025` | **P3** (Sinkhorn OT dispatch) | **MODEST / DEFERRED:** Modest PPL gain (`d=-2.663`); DoS & `>1B` FLOP savings deferred |
| Lookahead World Rollouts | `Ha & Schmidhuber 2018`, `Hafner 2023`, `Bardes 2021` | **P4** ($z_{t+k} = T_\omega(z_t, a)$) | **BACKBONE-AGNOSTIC:** `100%` `K=5` vs `22–46%` `K=1`, but `TF+P4 = QNMF+P4` (`d=0.000`) |
| Execution Integrity & Abstention | Qadav Inventions `#11`, `#15`, `#16` | **P5** (`QERA` + `TTPCR` + `QEEH`) | **IMPLEMENTED (SYS) / TIED (ML):** Native `0xDEADBEEF` tamper fail-stop passes 100%; learned verifier ties (`93.82%` prec.) |
| Dynamic Compute Budgeting | `Graves 2016`, `Kapusuzoglu 2026` | **P6** (ECRA 4-tier routing) | **PARETO TRADE-OFF:** `54.88%` step savings at `-3%` M9.14 / `-4%` M9.15 raw accuracy cost |

---

## 3. QNMF Mathematical Architecture (M9.10) & Native Systems Foundation

### 3.1 Formal Equations of the Six QNMF Pillars

Formalized in [`docs/QNMF_MATH_SPEC.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_MATH_SPEC.md), QNMF models a discrete step $t \in \mathbb{N}$ mapping token/sensory input $x_t \in \mathbb{R}^{d_{\text{in}}}$ through latent causal state $S_t \in \mathbb{R}^{d_s}$, three memory manifolds $(M_s^{(t)}, M_w^{(t)}, M_l^{(t)})$, expert simplex routing $r_t \in \Delta^{K-1}$, world-model state $z_t \in \mathbb{R}^{d_z}$, cryptographic proof certificate $\kappa_t \in \mathcal{K}$, and deliberation budget $B_t \in \mathbb{R}_{\ge 0}$:

1. **Pillar 1 — Commutative Bilateral Causal State Dynamics ($S_t$):**
   To mitigate directional asymmetry in 1D causal convolutions, P1 applies a commutative symmetric projection $\Phi(x_t)$ with involution $\mathcal{T}$ and $\text{Sym}(M) = \frac{1}{2}(M + M^\top)$ inside a strictly damped diagonal state update ($\Lambda_t \in (0, 1)^{d_s}$):
   $$\Delta_t = \text{softplus}(w_\Delta^\top x_t + b_\Delta), \qquad \Lambda_t = \exp\!\left(-\Delta_t \cdot \text{diag}(\alpha_1, \dots, \alpha_{d_s})\right)$$
   $$\Phi(x_t) = \frac{1}{2}\left(W_u x_t + W_v \mathcal{T}(x_t)\right) + \text{Sym}(W_a x_t \otimes W_b x_t) v_{\text{proj}}$$
   $$S_{t+1} = \Lambda_t \odot S_t + (1 - \Lambda_t) \odot \Phi(x_t) + W_{\text{mem}}\,\text{Read}(M_w^{(t)}, S_t) + \sum_{k=1}^{K} r_{t, k} E_k(S_t)$$
   Trained with spectral stability regularization $\mathcal{R}_{\text{spectral}}(\Lambda) = \frac{1}{d_s}\sum_i \max(0, \Lambda_{t,i} - (1 - 10^{-4}))^2$.

2. **Pillar 2 — Multi-Timescale Hierarchical Memory ($M_s \to M_w \to M_l$):**
   * *Short-Term ($M_s \in \mathbb{R}^{N_s \times d_m}$):* FIFO ring buffer retaining recent activations.
   * *Working Associative Memory ($M_w \in \mathbb{R}^{N_w \times d_m}$):* Surprise-gated ($\sigma_t = \|x_t - \text{Predict}(S_t)\|_2^2 > \tau_{\text{surprise}}$) normalized outer-product delta write:
     $$M_w^{(t+1)} = (1 - \gamma_t) M_w^{(t)} + \eta_t \cdot \frac{k_t \otimes (v_t - M_w^{(t)} k_t)}{1 + \|k_t\|_2^2}$$
     where $\gamma_t = \text{sigmoid}(w_\gamma^\top S_t)$ and $\eta_t = \text{sigmoid}(w_\eta^\top S_t)$.
   * *Crystalline Long-Term Store ($M_l$, key-value $(K_l, V_l)$):* Keyed normalized slots enabling $O(1)$ slot invalidation ($\text{Revoke}(h_k)$ zeroing $K_{l,j}, V_{l,j}$ without gradient passes across $\theta_{\text{slow}}$).

3. **Pillar 3 — Entropic Optimal Transport Cognitive Routing ($r_t$):**
   Given batch size $B$ and $K$ experts with centroids $\mu_k$ and dispatch cost $C_{i,k} = \frac{1}{2}\|W_{\text{gate}} x_i - \mu_k\|_2^2$, Sinkhorn-Knopp iterations compute the transport plan $P^* \in \Pi(u, v)$ solving $\min_{P \in \Pi(u,v)} \langle P, C \rangle - \varepsilon \mathcal{H}(P)$:
   $$P^* = \text{diag}(u^{(\tau)}) \exp\!\left(-\frac{C}{\varepsilon}\right) \text{diag}(v^{(\tau)}), \qquad r_i = \text{TopK}(P^*_{i,:}, k_{\text{active}})$$

4. **Pillar 4 — Predictive Latent World Model ($T_\omega$):**
   Transition dynamics unfold over compact latent representations $z_t = g_\omega(S_t) \in \mathbb{R}^{d_z}$ with action space $\mathcal{A}$ (`UP`, `DOWN`, `LEFT`, `RIGHT` in M9.13):
   $$\hat{z}_{t+1} = T_\omega(z_t, a_t) = z_t + \Delta \tau \cdot \mathcal{F}_\omega(z_t, a_t), \qquad \mathcal{L}_{\text{P4}}(\omega) = \sum_{k=1}^{H} \gamma^k \|z_{t+k} - \hat{z}_{t+k}\|_2^2 + \lambda_{\text{var}}\mathcal{V}(Z) + \lambda_{\text{cov}}\mathcal{C}(Z)$$
   Lookahead planning unrolls $K$-step trajectories over learned reward/value predictions (`rollout_search`) rather than token sampling.

5. **Pillar 5 — Cryptographic Evidence-Guided Reasoning & Fail-Stop ($V(\kappa_t)$):**
   Steps bind premise digests, operators, witness pointers $w_t$, and calibrated confidence $c_t \in [0, 1]$ into SHA-256 / SHA3-256 chain records verified against `DECOHERE` / `REFUSE` / `ACCEPT`:
   $$V(\kappa_t, h_{\text{pre}}, h_{\text{concl}}) = \begin{cases} \text{ACCEPT} & \text{if digest matches and } c_t \ge \tau_{\text{safe}} \\ \text{REFUSE} & \text{if } c_t < \tau_{\text{safe}} \\ \text{DECOHERE} & \text{if cryptographic digest mismatch} \end{cases}$$

6. **Pillar 6 — Evidence-Calibrated Adaptive Compute Allocation / ECRA ($B_t$):**
   A lightweight learned router scores feature/uncertainty vectors to assign discrete compute levels (`1, 2, 4, 8` steps) backed by ACT ponder cost regularization $\mathcal{L}_{\text{P6}} = \mathcal{L}_{\text{task}} + \lambda_{\text{ponder}} \sum (N_t + R_t)$ and post-execution verification thresholding.

### 3.2 Native Zero-Libc Implementation & Attribution

Notice two essential scientific disclosures regarding novelty and implementation:
* **Explicit Attribution of Borrowed Mathematical Primitives:** State-space exponential recurrence draws on Kalman filtering and `S4`/`Mamba`/`RWKV`; Sinkhorn transport draws on Cuturi (2013); VICReg covariance/variance losses come from `Bardes et al. (2021)`; and ACT halting draws on `Graves (2016)`. QNMF's candidate architectural contribution lies specifically in unifying commutative bilateral dynamics, multi-timescale slotted memory, Sinkhorn dispatch, latent rollout search, and **native cryptographic witness predicates (`P5`) inside a zero-libc compiler runtime**—not in inventing individual optimization building blocks.
* **Freestanding Native Engine (`internal/runtime/qadav_nn.c`):** Rather than shelling out to Python/PyTorch, every M9.11–M9.15 model trains and evaluates inside compiled Windows binaries linked purely against `kernel32.lib`, using `VirtualAlloc` (`nn_large_alloc`), custom Taylor/Padé transcendental approximations (`qadav_exp`, `qadav_log`, `qadav_sqrt`, `fast_tanh`, `fast_sigmoid`, `fast_exp`), deterministic `xorshift64` PRNG streams (`nn_seed_rng`), and hand-written AdamW gradient steps clipped to `[-5.0, +5.0]`. Exact compiled parameter allocations observed in `internal/runtime/qadav_nn.c` are:
  * **M9.11 Toy (`TOY_*`):** `QNMF_PARAM_COUNT = 309,572` (`0.31M`), `TRANSFORMER_PARAM_COUNT = 421,184` (`0.42M`), `NOROUTING_PARAM_COUNT = 111,296` (`0.11M`), vocab `64`, emb `64`, state `128`, `4` experts (`hidden=256`).
  * **M9.12 Scaled v2 (`V2_*`):** `QNMF_V2_PARAM_COUNT = 17,557,648` (`17.56M`), `TRANSFORMER_V2_PARAM_COUNT = 17,671,040` (`17.67M`), `QNMF_V2_NO_ML_PARAM_COUNT = 17,492,112` (`17.49M`), `QNMF_V2_NO_MW_PARAM_COUNT = 17,541,264` (`17.54M`), vocab `128`, emb `256`, state `256`, `16` experts (`hidden=2048`), `M_w=64` slots, `M_l=256` slots.
  * **M9.13 World Model:** `V3_WM_PARAM_COUNT = 92,418`, latent $d_z = 128$, `4` discrete actions, `8x8` hard grid (`18` walls forming `4` U-shaped traps).
  * **M9.14 ECRA:** `EcraRouter` MLP `129 -> 64 -> 4`, paired with `EcraVerifier` operating over `16` extracted trace/confidence features.

### 3.3 Systems Invariants (`M9.1–M9.8`, `M9.4`, `M9.4a`)

Before M9.9–M9.15 neural ablations, Track I verified 8 security/cryptographic runtime inventions inside standalone `612–614 KB` native binaries compiled by `bin/qyn.exe` and cross-checked by the Phase 8 regression suite (`10/10 fc.exe /b` identical LLVM IR output). Key milestone repairs completed during Phase 9 include:
* **`M9.4` BARAT Frobenius Norm Fix:** Initial `product_barat` tests emitted an erroneous outlier norm (`0.06`). Correcting the C runtime reduction over `grad_w3 = [[100.0, 100.0], [100.0, 100.0]]` restored exact Euclidean math: $\|\nabla w_3\|_F = \sqrt{4 \times 100^2} = 200.000000$.
* **`M9.4a` Compiler 2D Tensor Literal Emission:** `selfhost/codegen.ydv` previously failed to lower native 2D matrix literals accurately in LLVM IR, forcing hardcoded workarounds in `qadav_barat.c`. Fixing IR emission directly in the self-hosted compiler removed the C workaround, allowing `product_barat.exe` and `product_integrated.exe` (`87 ms`, 9 chained steps from MCTP provenance through QERA audit root) to execute genuine `.ydv`-defined 2D tensors.
* **`QEEH` Tamper Containment:** Across happy vs. tamper pairs (`product_qemwv`, `product_qera`, `product_ttpcr`, `product_barat`, `product_cdaga`, `product_qecma`, `product_qesi`, `product_integrated`), bit flips, tag swaps, and certificate forgeries reliably trigger buffer zeroization and `ExitProcess(0xDEADBEEF)` (`-559038737` signed), while calibrated abstention (`product_ttpcr_abstain.exe`, confidence `0.55 < 0.80`) and contained agent sandbox crashes (`product_qesi_isolation.exe`) preserve clean host `Exit Code: 0`.

---

## 4. Experimental Methodology, Controls & Hardware Telemetry

### 4.1 Hardware Profile & Memory Envelope

All benchmarks were executed on a **single Windows x86_64 consumer host (8 GB RAM target class, CPU-only, zero GPU/CUDA/WebGPU acceleration)**. Peak runtime resource telemetry measured live during `M9.15` evaluation established:
* **Peak Live Working Set (RSS):** **`540.32 MB`**
* **Peak Private Commit:** **`546.25 MB`**
* **Aggregate `M9.15` Benchmark Wall-Clock Time (`63` runs):** **`140.69 s`** (`93.79 s` normalized per 42 runs), decomposed strictly across tasks as **`103.79 s`** (Language Modeling: `21` evaluation runs), **`36.46 s`** (Hard Grid World Model: `21` rollout runs), and **`0.40 s`** (ECRA: `21` routing/verification runs).
* **Pretraining Time (`M9.12`, 2,000 steps @ batch 4, ctx 32):** **`3,886.7 s`** (`~64.8 min`) for `QNMF v2` (`17.56M`) vs. **`402.1 s`** (`~6.7 min`) for `Transformer v2` (`17.67M`).

### 4.2 Why Earlier Revision Artifacts Were Rejected & Replaced

To enforce rigorous reproducibility, any earlier Phase 9 artifact suffering from undertraining, degenerate benchmarks, asymmetric controls, or uncalibrated variance was discarded and rebuilt:
1. **`M9.11` & `M9.12` (Undertrained 20-step runs replaced):** Early runs stopped at `20` steps (oscillating cross-entropy `3.61 -> 3.89`), making perplexity comparisons noise-dominated. Adding `--steps` and running **`2,000` full Adam steps** produced monotone smooth convergence trajectories (`TMP/qnmf_loss.csv`, `TMP/qnmf_v2_loss.csv`), while additionally replacing `.ydv` wrapper shortcuts with honest C ablation/driver paths (`ablation_driver.exe`).
2. **`M9.13` (Trivial open grid + asymmetric baseline replaced):** An initial open `8x8` grid let myopic `K=1` reach `100%`, hiding whether world-model rollouts were exercised. Introducing **4 U-shaped obstacle traps (`18` walls)** separated `K=5` (`100%`) from `K=1` (`46%` QNMF, `22%` TF). Crucially, adding a fair **`Transformer + P4` (`92,418`-param world model) baseline** revealed that P4 works equally well atop Transformer encodings (`d = 0.000`).
3. **`M9.14` (Over-idealized 0.0000 loss replaced):** Purely separable synthetic tiers collapsed router cross-entropy to `0.0000` (`100%` accuracy). Injecting **`20%` borderline task overlap, `4%` label noise, feature jitter (`sd = 0.085`), and weight decay (`0.0042`)** produced a realistic `0.1265` loss floor, `85.0%` tier accuracy, and genuine accuracy/compute trade-offs.
4. **`M9.15` (`N=2` artifacts & understated timing replaced):** Replaced `N=2` variance artifacts, rigged planner asymmetries, and misreported fast wall times with **real `63`-run multi-seed execution (`42, 123, 999`)**, exact Student-$t$ 95% confidence intervals ($\nu = 4, t_{\text{crit}} = 2.7764$), real `140.69 s` timing, and explicit `INCONCLUSIVE` / low-variance footnotes.

### 4.3 Controlled Variant Isolation (`M9.15` 7-Model Matrix)

In `M9.15` (`internal/runtime/qadav_nn.c`, `qadav_run_ablation_driver`), `7` model configurations were evaluated across `3` tasks (`lm`, `grid`, `ecra`) and `3` seeds (`42, 123, 999` -> `63` CSVs in `TMP/abl_*.csv`):
* `transformer`: `17.67M` Transformer v2 (`lm`), **`Transformer+P4` world model (`grid`)**, and `Fixed L3 (8 steps, unverified)` (`ecra`).
* `qnmf_full`: `17.56M` QNMF v2 (`lm`), `QNMF+P4 K=5` (`grid`), and `ECRA + QNMF Verifier` (`ecra`).
* `qnmf_P1`: Ablated P1 variant using recurrent state/routing configuration for `lm`, sharing full `grid` (`K=5`) and `ecra` paths.
* `qnmf_P2`: `17.54M` `no-Mw` checkpoint (`qnmf_v2_no_mw.bin`, disabling plastic working memory $M_w$) on `lm`, sharing `grid` and `ecra` paths.
* `qnmf_P3`: Sinkhorn sparse routing ablated (`is_p3=1`) on `lm`, sharing `grid` and `ecra` paths.
* `qnmf_P4`: Disables world-model planning (`random` policy) on `grid`, while retaining `qnmf_full` weights on `lm` and `ECRA` on `ecra`.
* `qnmf_P6`: Disables ECRA dynamic routing (forcing Fixed L3 `8.00` steps) on `ecra`, while retaining `qnmf_full` weights on `lm` and `K=5` planning on `grid`.

Because `qnmf_P4` and `qnmf_P6` intentionally alter only their target subsystems (`grid` and `ecra`, respectively), their Task 1 LM perplexities equal `qnmf_full` by construction (`d = 0.000`). Similarly, Task 2 grid rollouts test the P4 planner specifically (`qnmf_full` vs `transformer+P4` vs `qnmf_P4`), and Task 3 compares dynamic ECRA (`qnmf_full`) against Fixed L3 (`transformer`/`qnmf_P6`).

---

## 5. Empirical Results Across Milestones M9.11–M9.15

### 5.1 M9.11 — Toy Sequence Modeling (2,000 Steps, Character Shakespeare)

Milestone `M9.11` compares toy QNMF (`309,572` params) against toy Transformer (`421,184` params) on the first `900,000` bytes of `TMP/shakespeare_subset.txt` (`TOY_BATCH_SIZE=32`, `TOY_CONTEXT_LEN=64`, `TOY_VOCAB_SIZE=64`, Adam `lr=0.003`) for `2,000` optimization steps, evaluated on the `100,000`-byte held-out split (`val_offset=900000`, `seed=777`). Both models train smoothly over `2,000` steps without collapse (`TMP/qnmf_loss.csv`, `TMP/transformer_loss.csv`):

* **QNMF Toy (`0.31M` params):** Step `0`: `4.133323` -> Step `500`: `2.054566` -> Step `1000`: `1.884245` -> Step `1500`: `1.703165` -> Step `2000`: `1.613953` (held-out val PPL **`6.10`**).
* **Transformer Toy (`0.42M` params):** Step `0`: `4.197031` -> Step `500`: `2.658486` -> Step `1000`: `2.515004` -> Step `1500`: `2.388673` -> Step `2000`: `2.430268` (held-out val PPL **`12.72`**).

Note that `M9.11` toy `.ydv` wrappers call `qadav_nn.c` directly rather than building deep matrix layers natively in `.ydv` loop syntax; this engineering boundary was explicitly logged as an honest stub/wrapper boundary prior to M9.15.

### 5.2 M9.12 — Scaled 17.6M Evaluation, Speed Slowdown & Associative Recall Failure

Milestone `M9.12` scales both architectures into `17.49M–17.67M` parameters (`V2_EMB_DIM=256`, `V2_STATE_DIM=256`, `V2_NUM_EXPERTS=16`, `V2_EXPERT_HIDDEN=2048`, `V2_VOCAB_SIZE=128`) and trains all four variants for `2,000` steps (`batch=4`, `ctx=32`, Adam `lr=0.0020`).

#### Table 3: M9.12 17.6M Scaled Language Modeling & Associative Recall (`eval_v2.exe`)

| Model | Params | Ctx `64` PPL | Ctx `128` PPL | Ctx `256` PPL | Ctx `512` PPL | Step `0` -> `2000` Train Loss | Associative Recall (@500 tokens, 50 needles) | 2,000-Step CPU Train Time |
|---|---|---|---|---|---|---|---|---|
| **QNMF v2** | `17.56M` | **6.48** | **6.17** | **6.83** | **6.71** | `4.9759` -> `1.9109` | **10.0%** (`5/50` = random chance) | **`3,886.7 s`** (`9.67x` slower) |
| **Transformer v2** | `17.67M` | 17.29 | 17.40 | 17.28 | 17.01 | `5.1062` -> `2.6091` | **10.0%** (`5/50` = random chance) | **`402.1 s`** (**Baseline 1.00x**) |
| **QNMF v2 (no-Ml)** | `17.49M` | 128.00 | 128.00 | 128.00 | 128.00 | `4.8520` -> `4.8520` | **10.0%** (`5/50` = random chance) | — |
| **QNMF v2 (no-Mw)** | `17.54M` | 30.70 | 33.31 | 35.02 | 35.37 | `4.8983` -> `3.5548` | **10.0%** (`5/50` = random chance) | — |

Three unambiguous conclusions follow from Table 3:
1. **Convergence Gap at Step 2,000:** QNMF v2 reaches `6.71` perplexity at context `512`, outperforming Transformer v2 (`17.01`) under fixed `2,000`-step budget (`batch=4`, `ctx=32`). However, notice that `2,000 * 4 * 32 = 256,000` training token updates is far below Chinchilla-optimal compute (`20 * 17.6M ≈ 352M` tokens, `Hoffmann et al., 2022`), meaning `Transformer v2`'s slower early loss trajectory (`5.11 -> 2.61` vs. QNMF's `4.98 -> 1.91`) reflects **early-stage optimization under tiny CPU budgets**, not proven converged asymptotic superiority.
2. **Prohibitive Scalar CPU Slowdown (`9.67x`):** `QNMF v2` takes **`3,886.7 s`** versus **`402.1 s`** for `Transformer v2`—nearly an **order of magnitude slower** per 2,000 training steps. Inner sequential token loops updating $S_t$, `64`-slot $M_w$, `256`-slot $M_l$, and `16`-expert Sinkhorn dispatch cannot exploit contiguous tensor GEMM tiling in zero-libc scalar C (`-O0` / scalar double precision).
3. **Associative Recall Null Result (`10.0%` = Random Digit Chance):** Every single model—including `QNMF v2` with both `M_w` and `M_l` intact—scores identically at **`10.0%` (`5/50`)** on the 5 needle-in-a-haystack digits probe over 500 tokens. Short byte-window Shakespeare pretraining never teaches explicit key-value needle copying across distractors. **Proposition P1 / Protocol EXP-2 recall claims are therefore strictly unverified and deferred.**
4. **Memory Dependence:** Disabling long-term slotted state `M_l` leaves `QNMF v2 (no-Ml)` frozen at `log(128) = 4.852030` (perplexity `128.00`), while disabling working memory `M_w` degrades perplexity by `~5.3x` (`6.71 -> 35.37`, training loss `3.5548`). Both memory manifolds are load-bearing within QNMF's forward computation graph.

### 5.3 M9.13 — Hard 4-Trap World-Model Planning (100 Episodes)

To test Pillar 4 without trivial open-grid bias, `M9.13` evaluates 100 episodes (`env_grid.ydv` -> `TMP/plan_ep_*.csv`) across an **`8x8` maze containing `18` wall cells organized into `4` U-shaped dead-end traps** where Manhattan-greedy policies trap cleanly. Crucially, both QNMF representations and **Transformer encodings (`encode_grid_state_mode(..., is_tf=1)`)** were trained with equivalent `92,418`-parameter P4 transition/value world models (`32` multi-step `H=5` unrolled sequences per episode).

#### Table 4: M9.13 Hard 4-Trap `8x8` Grid Evaluation (100 Episodes per Policy)

| Policy Configuration | Lookahead Horizon | Success Rate (%) | Mean Steps (All / Succ) | Interpretation |
|---|---|---|---|---|
| **QNMF + P4** | `K = 5` | **100.0%** (`100/100`) | `9.66` / `9.67` | Avoids all 4 U-traps via latent rollouts |
| **Transformer + P4** | `K = 5` | **100.0%** (`100/100`) | `9.66` / `9.67` | **Matches QNMF+P4 identically (`d = 0.000`)** |
| **QNMF (Myopic)** | `K = 1` | **46.0%** (`46/100`) | `~10` (succ) / `100` (fail) | Fails inside deeper U-trap basins |
| **Transformer (Myopic)** | `K = 1` | **22.0%** (`22/100`) | `~10` (succ) / `100` (fail) | Severe greedy myopic trap capture |
| **Transformer (Random)** | `0` (No WM) | **17.0%** (`17/100`) | — | Random action walk baseline |
| **QNMF (Random)** | `0` (No WM) | **14.0%** (`14/100`) | — | Random action walk baseline |
| **QNMF (No Reward)** | `K = 5` (Zeroed R) | **0.0%** (`0/100`) | `100.0` (timeout) | Complete failure without value signal |

Table 4 proves that **`K=5` world-model rollout depth enables solving the 4-trap maze (`100%` vs `22–46%` at `K=1`), AND proves that this capability belongs to the P4 world-model planner rather than the QNMF sequence backbone (`Transformer+P4 = QNMF+P4`)**.

### 5.4 M9.14 — Ambiguous ECRA Dynamic Routing & Learned Verification

To eliminate synthetic `100%` classification ceilings, M9.14 constructs `100` reasoning tasks stratified into 4 difficulty tiers (`30` Easy, `30` Medium, `20` Hard, `20` Very Hard) with **`20%` borderline tier ambiguity**, **`4%` label noise**, input feature jitter (`sd = 0.085`), uncertainty jitter (`sd = 0.03`), and weight decay (`0.0042`). Training `EcraRouter` (`129 -> 64 -> 4`) over `200` steps drives clean loss from `1.3976` down to a realistic **`0.1265` floor** (`85.0%` tier accuracy; `90.0%` true-difficulty accuracy), yielding the `TMP/ecra_routing_histogram.csv` confusion matrix:

#### Table 5: M9.14 Empirical Router Confusion Matrix (`TMP/ecra_routing_histogram.csv`, `n=100` Tasks)

| Ground-Truth Tier | Count | Level `0` (`1` step) | Level `1` (`2` steps) | Level `2` (`4` steps) | Level `3` (`8` steps) | Tier Exact Acc |
|---|---|---|---|---|---|---|
| Tier `0` (Easy) | `30` | **25** | `2` | `1` | `2` | `83.3%` |
| Tier `1` (Medium) | `30` | `2` | **25** | `2` | `1` | `83.3%` |
| Tier `2` (Hard) | `20` | `0` | `0` | **17** | `3` | `85.0%` |
| Tier `3` (Very Hard) | `20` | `0` | `0` | `2` | **18** | `90.0%` |
| **Total / Weighted** | **100** | `27` | `27` | `22` | `24` | **85.0%** |

Evaluating compute efficiency vs. accuracy across these 100 ambiguous tasks gives M9.14's defining Pareto trade-off:
* **Compute Consumption:** ECRA allocates **`361` total steps** (`27*1 + 27*2 + 22*4 + 24*8`) vs. **`800` steps** (`100*8`) for Fixed Level `3` — **`54.88%` compute savings** (`0.4512` ratio).
* **Raw Accuracy Cost:** **`QNMF` trades `3.0%` raw accuracy for `54.88%` compute savings** (`82.0%` raw accuracy under ECRA vs. `85.0%` Fixed Level 3 in M9.14 single evaluation).
* **Learned Verifier & Abstention (`P5` Integration):** Enabling the learned `16`-feature verifier achieves **`76.0%` verified correct, `19.0%` calibrated abstention, and `5.0%` uncaught errors** (reducing erroneous outputs vs. `15.0%` unverified Fixed L3). Crucially, **both QNMF (`tau = 0.66`) and Transformer (`tau = 0.50`) learned verifiers achieve identical discriminative metrics: Precision = `93.82%` (`76/81`) and Recall = `72.22%` (`13/18`)**, confirming that verification selectivity derives from the P5 trace-scoring mechanism regardless of underlying sequence model.

### 5.5 M9.15 — Complete 7-Model × 3-Task × 3-Seed Ablation Synthesis

Running `ablation_driver.exe` across `3` seeds (`42, 123, 999`) for all `7` models on Tasks 1 (`lm`, `--batch 2`), 2 (`grid`, `--episodes 8` on M9.13 hard traps), and 3 (`ecra`, `100` tasks) generated `63` raw CSVs (`TMP/abl_*.csv`, complete values in Appendix A) aggregated in `TMP/ablation_summary.csv` and reproduced in Table 6 (`140.69 s` total real CPU time, `540.32 MB` peak RSS).

#### Table 6: M9.15 Multi-Seed Ablation Summary (Mean ± Std across `N=3` Seeds: `42, 123, 999`)

| Model Config | LM PPL `64` | LM PPL `128` | LM PPL `256` | LM PPL `512` | Grid Succ (%) | Grid Steps | ECRA Compute Ratio | ECRA Raw Acc (%) | ECRA Verif Acc (%) | ECRA Wrong (%) |
|---|---|---|---|---|---|---|---|---|---|---|
| **Transformer** (`TF+P4` / Fixed L3) | `16.09 ± 0.75` | `15.81 ± 2.91` | `16.01 ± 0.44` | `18.66 ± 0.37` | **100.0 ± 0.0** | `9.79 ± 1.56` | `1.0000 ± 0.0000` | **84.7 ± 3.5** | `84.7 ± 3.5` | `15.3 ± 3.5` |
| **QNMF Full** (`P1..P6`) | **7.52 ± 0.13** | **7.21 ± 1.03** | **7.61 ± 0.59** | **7.57 ± 0.81** | **100.0 ± 0.0** | `9.79 ± 1.56` | **0.4512 ± 0.0000** | `80.7 ± 1.2` | `75.3 ± 0.6` | **4.7 ± 2.5** |
| **QNMF-P1** (P1 ablated) | `14.53 ± 0.47` | `13.60 ± 2.12` | `13.82 ± 0.28` | `14.87 ± 0.39` | `100.0 ± 0.0` | `9.79 ± 1.56` | `0.4512 ± 0.0000` | `80.7 ± 1.2` | `75.3 ± 0.6` | `4.7 ± 2.5` |
| **QNMF-P2** (No $M_w$) | `39.19 ± 2.59` | `36.32 ± 2.82` | `35.60 ± 2.11` | `37.45 ± 1.58` | `100.0 ± 0.0` | `9.79 ± 1.56` | `0.4512 ± 0.0000` | `80.7 ± 1.2` | `75.3 ± 0.6` | `4.7 ± 2.5` |
| **QNMF-P3** (No Sinkhorn) | `9.12 ± 0.53` | `8.79 ± 1.30` | `9.09 ± 0.43` | `9.29 ± 0.43` | `100.0 ± 0.0` | `9.79 ± 1.56` | `0.4512 ± 0.0000` | `80.7 ± 1.2` | `75.3 ± 0.6` | `4.7 ± 2.5` |
| **QNMF-P4** (No WM / Random) | `7.52 ± 0.13` | `7.21 ± 1.03` | `7.61 ± 0.59` | `7.57 ± 0.81` | **8.3 ± 7.2** | `81.67 ± 30.04` | `0.4512 ± 0.0000` | `80.7 ± 1.2` | `75.3 ± 0.6` | `4.7 ± 2.5` |
| **QNMF-P6** (Fixed L3, No ECRA) | `7.52 ± 0.13` | `7.21 ± 1.03` | `7.61 ± 0.59` | `7.57 ± 0.81` | `100.0 ± 0.0` | `9.79 ± 1.56` | **1.0000 ± 0.0000** | **84.7 ± 3.5** | `84.7 ± 3.5` | `15.3 ± 3.5` |

---

## 6. Statistical Analysis, Cohen's $d$ & Proposition Scorecard

### 6.1 Effect Sizes (`d`) & 95% Student-$t$ Confidence Intervals ($\nu = 4, t_{\text{crit}} = 2.7764$)

Because each Task 1 evaluation runs over a fixed checkpoint and deterministic held-out window selection (`--seed` perturbs stochastic sampling masks/episode order rather than full weight re-initialization from scratch), standard deviations across seeds are tight (`0.37–1.58` PPL). Computing pooled Cohen's $d = (\bar{X}_A - \bar{X}_B)/s_{\text{pooled}}$ on tiny $s_{\text{pooled}}$ mechanically inflates `|d| > 10`. **These values are explicitly flagged with an asterisk (`*`) as artifacts of low seed variance across `N=3` fixed-checkpoint runs rather than universal population certainty.** Full training-from-scratch multi-seed variance has not yet been measured at `17.6M` scale due to the `~65 min/run` CPU training cost.

#### Table 7: Pairwise Statistical Comparisons (M9.15 Revision Verified)

| Task & Metric | Comparison Pair | Mean Difference | Cohen's $d$ | 95% Student-$t$ CI ($\nu=4$) | Statistical & Scientific Interpretation |
|---|---|---|---|---|---|
| **LM PPL @ `512`** | `QNMF Full` vs. `Transformer` | **-11.09** | `-17.659*` | `[-12.51, -9.67]` | Lower checkpoint PPL at 2,000 steps (`*` = fixed-ckpt low variance; `9.67x` slower CPU training) |
| **LM PPL @ `512`** | `QNMF Full` vs. `QNMF-P1` | **-7.30** | `-11.510*` | `[-8.73, -5.86]` | Confirms P1 recurrent dynamics/ablated path shifts LM convergence (`*` low seed variance) |
| **LM PPL @ `512`** | `QNMF Full` vs. `QNMF-P2` (no $M_w$) | **-29.88** | `-23.859*` | `[-32.72, -27.04]` | Largest measured LM degradation; working memory $M_w$ is essential (`*` low seed variance) |
| **LM PPL @ `512`** | `QNMF Full` vs. `QNMF-P3` (no Sinkhorn) | **-1.72** | `-2.663` | `[-3.19, -0.26]` | **Modest** statistically separable contribution; not a large-FLOP proof on 16 experts |
| **LM PPL @ `512`** | `QNMF Full` vs. `QNMF-P4` / `QNMF-P6` | **0.00** | **`0.000`** | `[-1.83, +1.83]` | **`INCONCLUSIVE` / Neutral:** P4/P6 subsystems are orthogonal to passive LM perplexity |
| **Grid Succ (%)** | `QNMF+P4` vs. **`Transformer+P4`** | **0.0%** | **`0.000`** | `[0.00, 0.00]` | **Equivalence:** Both `100.0%` (`9.79` steps); **P4 is backbone-agnostic** |
| **Grid Succ (%)** | `QNMF+P4` vs. `QNMF-P4` (Random) | **+91.7%** | `+17.963*` | `[+80.10, +103.24]` | Lookahead rollouts solve hard U-traps that random actions fail (`*` deterministic ceiling) |
| **Grid Succ (%)** | `QNMF+P4` vs. `QNMF (No Reward)` | **+100.0%** | `+infinity` | Exact binary `[100% vs 0%]` | Value signal is strictly necessary for directed search |
| **ECRA Ratio** | `QNMF Full` vs. `TF / P6 Fixed L3` | **-0.5488** | `-infinity` | Exact binary `[0.4512 vs 1.0]` | Deterministically routes **54.88% fewer compute steps** |
| **ECRA Raw Acc (%)** | `QNMF Full` vs. `TF / P6 Fixed L3` | **-4.00%** | **`-1.530`** | `[-9.93%, +1.93%]` | **`INCONCLUSIVE` (CI spans `0`); trades ~3%–4% raw accuracy for compute savings** |
| **ECRA Wrong (%)** | `QNMF Full` vs. `TF / P6 Fixed L3` | **-10.67%** | `-3.491` | `[-17.59%, -3.74%]` | Statistically significant error drop via `20%` abstention & learned verifier |

### 6.2 Assessment Against M9.10 Mathematical Propositions

Table 8 audits each theoretical proposition in [`docs/QNMF_MATH_SPEC.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_MATH_SPEC.md) strictly against Phase 9 empirical outcomes, prohibiting speculative extrapolations beyond measured evidence.

#### Table 8: M9.10 Mathematical Propositions & Falsifiable Predictions Evaluation

| Proposition / Prediction | Original M9.10 Falsifiable Statement | Observed Phase 9 Reality | Honest Status |
|---|---|---|---|
| **Prop. 1 / EXP-2:** Sub-Quadratic Scaling & `1M` Recall | Flat $O(1)$ per-token memory; $\ge 95\%$ recall on `1M` needles | Memory allocations stay bounded (`540.32 MB` peak RSS up to `512` ctx), **but `500`-token associative recall = `10.0%` chance** and CPU step time is `9.67x` slower | **UNVERIFIED / DEFERRED** |
| **Prop. 2 / EXP-1 & EXP-6:** Outside $TC^0$ & Symmetry | $\ge 99\%$ reverse recall / arbitrary-length parity | Not benchmarked in Phase 9 (`EXP-1`, `EXP-6` deferred); Shakespeare `512`-ctx PPL converges (`7.57`), which does not test $TC^0$ limits | **NOT TESTED / DEFERRED** |
| **Prop. 3 / EXP-4:** Sinkhorn Routing & DoS Immunity | Load variance `< 0.02` under RepetitionCurse DoS; `<3%` overhead | **P3** improves LM perplexity modestly (`9.29 -> 7.57`, `d=-2.663`), **but neither RepetitionCurse DoS (`10^4` repeats) nor `>1B` FLOP speedups were run** | **PARTIAL (Modest LM effect; DoS & FLOPs deferred)** |
| **Prop. 4 / EXP-3 & EXP-5:** Plasticity & $O(1)$ Erasure | Zero adjacent forgetting under deletion / `50–100` sequential tasks | Ablating `M_w`/`M_l` breaks LM (`128.00` no-Ml, `37.45` no-Mw), proving memory dependence, **but `EXP-3` erasure & `EXP-5` continual stream tests remain unrun** | **PARTIAL (Dependence shown; Erasure/CL deferred)** |
| **Pred. P4.1:** Latent Rollouts vs. Myopic Planning | $\ge 82\%$ multi-step plan success vs $<40\%$ myopic | **`100.0%` success** on Hard 4-Trap grid (`K=5`) vs `22–46%` (`K=1`). **CRITICAL NUANCE:** `Transformer+P4` matches identically (`d=0.000`) | **VERIFIED FOR P4 (Backbone-agnostic)** |
| **Pred. P5.1 & P6.1:** Calibrated Abstention + ECRA Compute | $\ge 42\%$ FLOP reduction & uncertainty refusal without error spikes | **ECRA** cuts steps by **`54.88%`** (`-3%`/`-4%` raw acc cost) and cuts wrong-rate via `93.82%` precision verifier. Verifier works identically on TF (`tau=0.50`) | **PARTIAL (Pareto confirmed; ties TF verifier)** |

---

## 7. Architectural & Systems Discussion

### 7.1 Why Memory ($M_w, M_l$) Dominates Short-Budget Convergence while Failing Associative Recall

The sharp contrast between Table 3 (`QNMF v2` PPL `6.71` vs. `no-Mw` `35.37` and `no-Ml` `128.00`) and the universal **`10.0%` (`5/50`) needle recall score** illuminates an important structural pitfall in prototype neural memory design:
1. During `2,000` short-window character steps (`batch=4`, `ctx=32`), dense residual paths coupled through $M_l$ (`256` slots) and $M_w$ (`64` slots) furnish high-bandwidth parameterized projections that accelerate early n-gram density fitting relative to a vanilla `17.67M` Transformer trained without learning-rate warmup or multi-epoch scheduling.
2. However, because Shakespeare character prediction mostly rewards local `8–32` character statistical smoothness, gradient signals across `2,000` steps never penalize key-collision or force sparse content-addressable lookup across `500`-token spans. Consequently, **having architectural memory arrays does not automatically create associative recall behavior without an explicit retrieval training curriculum**—directly matching Yarden, Ben-Artzy & Schwartz (2026) on retrieval capacity failure modes.

### 7.2 Why QNMF Exhibits a `9.67x` CPU Training Bottleneck

Benchmarking inside zero-libc scalar C exposes raw algorithmic arithmetic intensity stripped of PyTorch/cuBLAS masking. A Transformer layer processes batches cleanly via dense projection matrices across context blocks, whereas `compute_qnmf_v2_batch` (`internal/runtime/qadav_nn.c`) executes token-by-token sequential dependency recurrences (`t = 0..T-1`), updating outer-product working buffers ($M_w$) and running iterative routing normalization inside scalar `double` loops (`402.1 s` TF vs. `3,886.7 s` QNMF). Scaling QNMF beyond `17.6M` CPU prototypes cannot rely on scalar CPU compilation; either AVX-512/SIMD micro-kernels or Qadav's planned WebGPU/CUDA compute backend (Phase 10+) is strictly necessary to make recurrent+Sinkhorn passes tractable.

### 7.3 Separating Neural Uncertainty from Deterministic Compiler Safety

A constructive finding spans both Track I and Track II of Phase 9: **cryptographic safety belongs in the runtime/compiler contract (`M9.1–M9.8`), whereas neural inference remains fundamentally probabilistic and fallible (`M9.11–M9.15`)**. When `product_qemwv.exe` or `product_integrated.exe` detects a flipped bit inside AES-GCM weight envelopes or SHA-256 QERA reasoning chains (`M9.7`), deterministic decoherence (`0xDEADBEEF`) enforces a hard zero-false-negative invariant regardless of model loss. Conversely, learned ECRA verification (`93.82%` precision, `72.22%` recall) still commits **`4.7 ± 2.5%` uncaught errors** and rejects valid solutions **`19%–23%`** of the time. Compelling AGI engineering keeps the deterministic `TTPCR`/`QERA`/`QESI` verification substrate independent of neural weights.

---

## 8. Threats to Validity & Explicit Limitations

To satisfy strict AGI engineering transparency, this section details every methodological limitation and unachieved objective of Phase 9:

### 8.1 Statistical & Experimental Threats to Validity
* **Fixed-Checkpoint `N=3` Seed Sampling (`M9.15`):** `ablation_driver.exe` evaluates `3` distinct evaluation/mask seeds (`42, 123, 999`) over converged checkpoints rather than executing `3 x 7 = 21` independent full `2,000`-step weight training runs from random initialization (`21 x 65 min` exceeds single-CPU feasibility). Accordingly, all `|d| > 10` values in Table 7 reflect **tight evaluation-window variance** and must not be read as universal architecture superiority.
* **Synthetic Grid & Task Construction (`M9.13`, `M9.14`):** Both the `8x8` 4-trap maze (`92,418`-param world model) and the `100`-task ECRA routing dataset (`20%` borderline tasks, `4%` noise) are controlled micro-benchmarks, not open-ended real-world codebases or web benchmarks.
* **Parameter Count & Step Horizon Mismatch (`M9.11`, `M9.12`):** Toy (`0.31M` vs. `0.42M`) and scaled (`17.56M` vs. `17.67M`) variants run only `2,000` Adam steps on a `900 KB` training subset. Standard `10,000+` step schedules or Chinchilla compute-matched comparisons (`>1B` tokens) could close or invert early perplexity curves.

### 8.2 Explicit Phase 9 Non-Deliverables (What Was NOT Demonstrated)
1. **No Competition with Frontier Models:** Zero benchmarks compare Qadav or QNMF against GPT-4 class, Claude, Gemini, or LLaMA-3 `8B–70B` models on standard suites (MMLU, GSM8K, HumanEval, NeedleBench).
2. **No Associative Recall Above Chance:** All 4 `17.6M` models score **`10.0%` (`5/50`)** at `500` tokens (`random digit chance`). Long-context key retrieval (`EXP-2`) failed to emerge organically and is deferred to synthetic curriculum training.
3. **No Multi-GPU Cluster or Hardware Enclave Verification:** Execution ran strictly on a single Windows x86_64 consumer CPU (`540.32 MB` peak RSS). WebGPU/CUDA kernels, distributed cluster BARAT consensus, and SGX/SEV hardware enclaves remain unimplemented.
4. **No Subword Tokenizer (`BPE` / `SentencePiece`):** M9.11/M9.15 operate on `64`-character ASCII vocabulary; M9.12 operates on `128`-byte characters; M9.13/M9.14 use synthetic vectors.
5. **No Converged Multi-Epoch Convergence (`10K+` steps):** Training stopped at `2,000` steps due to single-CPU time budgets (`64.8 min` per QNMF v2 run).
6. **No QNMF-Specific Planning Superiority (`d = 0.000`):** **`Transformer+P4 = QNMF+P4`** (`100.0 ± 0.0%` at `9.79 ± 1.56` steps). P4 planning is demonstrably backbone-agnostic.
7. **No Large-Scale Sparse Compute/DoS Speedups from P3:** On `4–16` tiny CPU experts, P3 Sinkhorn adds scalar runtime overhead while yielding only modest LM perplexity gains (`d = -2.663`). Demonstrating real FLOP reductions and `RepetitionCurse` immunity requires `>1B`-parameter models and fused GPU kernels.

---

## 9. Phase 10+ Experimental Roadmap

Every deferred claim in Table 8 maps directly to falsifiable Phase 10+ engineering criteria:
1. **Vectorized SIMD / WebGPU Compute Backend:** Replace scalar double loops inside `qadav_nn.c` with blocked matrix kernels and optional WebGPU shaders to bring QNMF step latency near dense matrix baselines (`< 2x` vs. current `9.67x` gap).
2. **Subword BPE Tokenizer & Synthetic Retrieval Curriculum (`EXP-1` / `EXP-2` / `EXP-3`):** Train a native zero-libc BPE tokenizer and re-run `17.6M–100M` models with explicit synthetic key-value, mirror-reversal, and $O(1)$ slot revocation curricula to test whether `P1`/`P2` break the `10.0%` recall floor when trained on retrieval supervision.
3. **Training-From-Scratch Multi-Seed Variance (`10,000+` Steps):** Once GPU acceleration lands, replicate `N >= 3` full independent training seeds up to Chinchilla saturation (`>= 20` tokens/param) so statistical intervals reflect true optimization variance rather than checkpoint evaluation variance.
4. **Formal Sparse Routing DoS Stress (`EXP-4`):** Run `10^4` repeated-token prompts across `K=64` experts to evaluate whether Sinkhorn transport preserves bounded load variance (`Var < 0.02`) against softmax balancing heuristics (`Fedus et al., 2022`; `Huang et al., 2025`).

---

## 10. Conclusion

Across Phase 9 revisions (`M9.1–M9.15`), Qadav transitioned from an unverified architectural proposal into an empirically tested native systems and cognitive prototype with clear boundaries. On the **systems axis**, Qadav compiles deterministic zero-libc binaries linking `45` C runtime objects, fixes prior compiler `2D` tensor emission (`M9.4a`) and gradient norm (`M9.4`, `200.0`) defects, passes `10/10` bitwise `fc /b` Phase 8 bootstrap regression checks, and enforces cryptographic weight, reasoning, provenance, multimodal, Byzantine, and sandbox fail-stops within `612–614 KB` executables. On the **cognitive ML axis**, rigorous post-revision benchmarks falsify oversimplified narratives: QNMF v2 (`17.56M`) achieves lower `2,000`-step Shakespeare perplexity (`6.71` vs. `17.01` M9.12; `7.57 ± 0.81` vs. `18.66 ± 0.37` M9.15) driven significantly by working/crystalline memory modules, **yet runs `9.67x` slower on CPU**, **remains at `10.0%` chance on associative needle recall**, shares **`d = 0.000` planning parity between `QNMF+P4` and `Transformer+P4` (`100.0%`)**, and **trades `3%–4%` raw accuracy for `54.88%` ECRA step savings** alongside `93.82%` precision learned verification. We conclude that QNMF warrants multi-seed GPU-cluster validation at `1B–7B` scale specifically on memory isolation (`P2`), Sinkhorn DoS stability (`P3`), and native safety predicates (`P5`), without claiming current superiority over scaled Transformers.

---

## Appendix A: Complete Raw Per-Seed Dataset (63 CPU Runs, `M9.15`)

All 63 CSV files (`TMP/abl_<model>_<task>_<seed>.csv`) recorded verbatim from `ablation_driver.exe` across `seeds = {42, 123, 999}` (`--batch 2` LM, `--episodes 8` Hard Grid, `100` ambiguous ECRA tasks):

### A.1 Task 1 — Language Modeling (`21` Runs)

| Model | Seed | PPL `64` | PPL `128` | PPL `256` | PPL `512` | Eval Loss (`512`) | Eval Time (`s`) |
|---|---|---|---|---|---|---|---|
| `transformer` | `42` | `15.25` | `12.55` | `16.50` | `18.25` | `2.9043` | `2.56` |
| `transformer` | `123` | `16.70` | `18.13` | `15.87` | `18.76` | `2.9322` | `2.43` |
| `transformer` | `999` | `16.32` | `16.75` | `15.65` | `18.97` | `2.9430` | `2.07` |
| `qnmf_full` | `42` | `7.40` | `6.43` | `6.93` | `7.29` | `1.9878` | `7.46` |
| `qnmf_full` | `123` | `7.65` | `8.38` | `7.88` | `8.48` | `2.1380` | `5.98` |
| `qnmf_full` | `999` | `7.51` | `6.81` | `8.02` | `6.94` | `1.9376` | `6.32` |
| `qnmf_P1` | `42` | `14.53` | `11.43` | `13.55` | `14.48` | `2.6729` | `5.70` |
| `qnmf_P1` | `123` | `14.06` | `15.67` | `13.79` | `15.26` | `2.7253` | `4.51` |
| `qnmf_P1` | `999` | `15.00` | `13.69` | `14.11` | `14.86` | `2.6989` | `4.59` |
| `qnmf_P2` | `42` | `36.20` | `34.33` | `33.25` | `38.67` | `3.6553` | `5.03` |
| `qnmf_P2` | `123` | `40.60` | `35.09` | `36.22` | `38.01` | `3.6379` | `5.03` |
| `qnmf_P2` | `999` | `40.78` | `39.55` | `37.32` | `35.67` | `3.5745` | `5.17` |
| `qnmf_P3` | `42` | `9.55` | `7.58` | `8.63` | `9.02` | `2.2001` | `4.84` |
| `qnmf_P3` | `123` | `9.29` | `10.16` | `9.14` | `9.79` | `2.2819` | `5.07` |
| `qnmf_P3` | `999` | `8.53` | `8.64` | `9.49` | `9.07` | `2.2050` | `5.48` |
| `qnmf_P4` | `42` | `7.40` | `6.43` | `6.93` | `7.29` | `1.9878` | `4.96` |
| `qnmf_P4` | `123` | `7.65` | `8.38` | `7.88` | `8.48` | `2.1380` | `5.54` |
| `qnmf_P4` | `999` | `7.51` | `6.81` | `8.02` | `6.94` | `1.9376` | `3.95` |
| `qnmf_P6` | `42` | `7.40` | `6.43` | `6.93` | `7.29` | `1.9878` | `4.17` |
| `qnmf_P6` | `123` | `7.65` | `8.38` | `7.88` | `8.48` | `2.1380` | `4.61` |
| `qnmf_P6` | `999` | `7.51` | `6.81` | `8.02` | `6.94` | `1.9376` | `4.92` |

### A.2 Task 2 — Hard 4-Trap Grid Planning (`21` Runs)

| Model | Seed | Success Rate (%) | Avg Steps | Mean Pred Error | Eval Time (`s`) |
|---|---|---|---|---|---|
| `transformer` (`TF+P4`) | `42` | `100.0` | `11.37` | `0.00000` | `2.17` |
| `transformer` (`TF+P4`) | `123` | `100.0` | `8.25` | `0.00000` | `1.46` |
| `transformer` (`TF+P4`) | `999` | `100.0` | `9.75` | `0.00000` | `1.81` |
| `qnmf_full` | `42` | `100.0` | `11.37` | `0.00005` | `2.48` |
| `qnmf_full` | `123` | `100.0` | `8.25` | `0.00004` | `1.64` |
| `qnmf_full` | `999` | `100.0` | `9.75` | `0.00005` | `1.95` |
| `qnmf_P1` | `42` | `100.0` | `11.37` | `0.00005` | `2.29` |
| `qnmf_P1` | `123` | `100.0` | `8.25` | `0.00004` | `1.50` |
| `qnmf_P1` | `999` | `100.0` | `9.75` | `0.00005` | `1.86` |
| `qnmf_P2` | `42` | `100.0` | `11.37` | `0.00005` | `2.54` |
| `qnmf_P2` | `123` | `100.0` | `8.25` | `0.00004` | `1.57` |
| `qnmf_P2` | `999` | `100.0` | `9.75` | `0.00005` | `1.82` |
| `qnmf_P3` | `42` | `100.0` | `11.37` | `0.00005` | `2.42` |
| `qnmf_P3` | `123` | `100.0` | `8.25` | `0.00004` | `1.71` |
| `qnmf_P3` | `999` | `100.0` | `9.75` | `0.00005` | `2.07` |
| `qnmf_P4` (Random) | `42` | `12.5` | `47.00` | `0.00002` | `0.03` |
| `qnmf_P4` (Random) | `123` | `0.0` | `100.00` | `0.00003` | `0.01` |
| `qnmf_P4` (Random) | `999` | `12.5` | `98.00` | `0.00002` | `0.01` |
| `qnmf_P6` | `42` | `100.0` | `11.37` | `0.00005` | `2.62` |
| `qnmf_P6` | `123` | `100.0` | `8.25` | `0.00004` | `1.81` |
| `qnmf_P6` | `999` | `100.0` | `9.75` | `0.00005` | `2.12` |

### A.3 Task 3 — Ambiguous ECRA (`21` Runs)

| Model | Seed | Avg Compute | Compute Ratio | Raw Acc (%) | Verified Acc (%) | Abstain (%) | Wrong (%) | Eval Time (`s`) |
|---|---|---|---|---|---|---|---|---|
| `transformer` (Fixed L3) | `42` | `8.00` | `1.0000` | `85.0` | `85.0` | `0.0` | `15.0` | `0.00` |
| `transformer` (Fixed L3) | `123` | `8.00` | `1.0000` | `81.0` | `81.0` | `0.0` | `19.0` | `0.00` |
| `transformer` (Fixed L3) | `999` | `8.00` | `1.0000` | `88.0` | `88.0` | `0.0` | `12.0` | `0.00` |
| `qnmf_full` | `42` | `3.60` | `0.4512` | `82.0` | `76.0` | `19.0` | `5.0` | `0.00` |
| `qnmf_full` | `123` | `3.60` | `0.4512` | `80.0` | `75.0` | `23.0` | `2.0` | `0.00` |
| `qnmf_full` | `999` | `3.60` | `0.4512` | `80.0` | `75.0` | `18.0` | `7.0` | `0.01` |
| `qnmf_P1` | `42` | `3.60` | `0.4512` | `82.0` | `76.0` | `19.0` | `5.0` | `0.00` |
| `qnmf_P1` | `123` | `3.60` | `0.4512` | `80.0` | `75.0` | `23.0` | `2.0` | `0.00` |
| `qnmf_P1` | `999` | `3.60` | `0.4512` | `80.0` | `75.0` | `18.0` | `7.0` | `0.00` |
| `qnmf_P2` | `42` | `3.60` | `0.4512` | `82.0` | `76.0` | `19.0` | `5.0` | `0.00` |
| `qnmf_P2` | `123` | `3.60` | `0.4512` | `80.0` | `75.0` | `23.0` | `2.0` | `0.00` |
| `qnmf_P2` | `999` | `3.60` | `0.4512` | `80.0` | `75.0` | `18.0` | `7.0` | `0.00` |
| `qnmf_P3` | `42` | `3.60` | `0.4512` | `82.0` | `76.0` | `19.0` | `5.0` | `0.00` |
| `qnmf_P3` | `123` | `3.60` | `0.4512` | `80.0` | `75.0` | `23.0` | `2.0` | `0.00` |
| `qnmf_P3` | `999` | `3.60` | `0.4512` | `80.0` | `75.0` | `18.0` | `7.0` | `0.00` |
| `qnmf_P4` | `42` | `3.60` | `0.4512` | `82.0` | `76.0` | `19.0` | `5.0` | `0.00` |
| `qnmf_P4` | `123` | `3.60` | `0.4512` | `80.0` | `75.0` | `23.0` | `2.0` | `0.00` |
| `qnmf_P4` | `999` | `3.60` | `0.4512` | `80.0` | `75.0` | `18.0` | `7.0` | `0.00` |
| `qnmf_P6` (Fixed L3) | `42` | `8.00` | `1.0000` | `85.0` | `85.0` | `0.0` | `15.0` | `0.01` |
| `qnmf_P6` (Fixed L3) | `123` | `8.00` | `1.0000` | `81.0` | `81.0` | `0.0` | `19.0` | `0.01` |
| `qnmf_P6` (Fixed L3) | `999` | `8.00` | `1.0000` | `88.0` | `88.0` | `0.0` | `12.0` | `0.00` |

---

## Appendix B: Hyperparameter & Architectural Constants (`internal/runtime/qadav_nn.c`)

| Subsystem | Constant / Parameter | Verified Value |
|---|---|---|
| **M9.11 Toy Models** | `TOY_VOCAB_SIZE` / `TOY_EMB_DIM` / `TOY_STATE_DIM` | `64` / `64` / `128` |
| **M9.11 Toy Experts & Optim** | `TOY_NUM_EXPERTS` / `TOY_EXPERT_HIDDEN` / `Batch` / `Ctx` | `4` / `256` / `32` / `64` |
| **M9.11 Toy Param Counts** | `QNMF_PARAM_COUNT` / `TRANSFORMER_PARAM_COUNT` / `NOROUTING` | `309,572` / `421,184` / `111,296` |
| **M9.11 Optimizer** | Adam `beta1=0.9`, `beta2=0.999`, `eps=1e-8`, Clip `[-5, +5]`, `lr` | `lr = 0.003` (`2,000` steps) |
| **M9.12 Scaled v2 Geometry** | `V2_VOCAB_SIZE` / `V2_EMB_DIM` / `V2_STATE_DIM` | `128` / `256` / `256` |
| **M9.12 Scaled Experts & Memory** | `V2_NUM_EXPERTS` / `V2_EXPERT_HIDDEN` / `V2_MW_SLOTS` / `V2_ML_SLOTS` | `16` / `2048` / `64` / `256` |
| **M9.12 Exact Parameter Counts** | `QNMF_V2` / `TRANSFORMER_V2` / `NO_ML` / `NO_MW` | `17,557,648` / `17,671,040` / `17,492,112` / `17,541,264` |
| **M9.12 Optimizer & Window** | Adam `lr` / Train `Batch` / `Ctx` / Steps / Held-out Offset | `0.0020` / `4` / `32` / `2,000` / `900,000` (`100 KB`) |
| **M9.13 World Model** | `V3_WM_PARAM_COUNT` / Latent $d_z$ / Rollout $H$ / Seqs per Ep | `92,418` / `128` / `5` / `32` (`8x8` hard `18`-wall grid) |
| **M9.13 Learning Rates & Seeds** | Base LR (`QNMF` / `TF`) / Train Seeds / Discount $\gamma$ | `0.0090` / `0.0055` / `42` / `142` / `0.98` |
| **M9.14 ECRA Router & Verifier** | Router `129 -> 64 -> 4` / LR / Ambiguity / Noise / `tau` (`QNMF`/`TF`) | `lr=0.015->0.009`, `wd=0.0042`, `20%` amb., `4%` noise / `0.66`/`0.50` |

---

## Appendix C: Reproducibility Commands & Artifact Manifest

To reproduce the `M9.12`, `M9.14`, `M9.15`, and Phase 8 bootstrap verifications from existing compiled binaries (zero source modifications):

```powershell
# Phase 8 Self-Hosted Fixed-Point Compiler Check (10/10 fc.exe /b must show no differences)
.\bin\qyn.exe emit-llvm examples\hello.ydv > TMP\p8_hello.ll
# ... repeated across all 10 Phase 8 proof examples against reference IR

# M9.12 17.6M Scaled Evaluation (Reads trained TMP\*.bin checkpoints)
.\TMP\eval_v2.exe

# M9.15 Single-Seed Ablation Driver Invocation Example (Generates TMP\abl_*.csv)
.\TMP\ablation_driver.exe --model qnmf_full --task lm --seed 42 --batch 2
.\TMP\ablation_driver.exe --model transformer --task grid --seed 42 --episodes 8
.\TMP\ablation_driver.exe --model qnmf_full --task ecra --seed 42
```

Primary artifact trails persisted in the repository: `docs/QNMF_PRIOR_ART_AUDIT.md` (M9.9), `docs/QNMF_MATH_SPEC.md` (M9.10), `TMP/qnmf_loss.csv` & `TMP/transformer_loss.csv` (M9.11), `TMP/qnmf_v2_loss.csv`, `TMP/transformer_v2_loss.csv`, `TMP/qnmf_v2_no_ml_loss.csv`, & `TMP/qnmf_v2_no_mw_loss.csv` (M9.12), `TMP/plan_ep_*.csv` (M9.13), `TMP/ecra_routing_histogram.csv` & `TMP/eval_ecra*.csv` (M9.14), `TMP/abl_*.csv`, `TMP/ablation_summary.csv` & `docs/QNMF_ABLATION_RESULTS.md` (M9.15).

---

## Appendix D: Live-Verified Bibliography

All 27 URLs listed below were live HTTP `200 OK` verified in this session (`November 11, 2026`); three unverified entries from earlier pre-revision drafts have been permanently removed.

1. **Bardes, A., Ponce, J., & LeCun, Y. (2021/2022).** *VICReg: Variance-Invariance-Covariance Regularization for Self-Supervised Learning.* arXiv:2105.04906. https://arxiv.org/abs/2105.04906
2. **Behrouz, A., Zhong, P., & Mirrokni, V. (2025).** *Titans: Learning to Memorize at Test Time.* arXiv:2501.00663. https://arxiv.org/abs/2501.00663
3. **Berglund, L., Tong, M., Kaufmann, M., Balesni, M., Stickland, A. C., Korbak, T., & Evans, O. (2023/2024).** *The Reversal Curse: LLMs trained on "A is B" fail to learn "B is A".* ICLR 2024 / arXiv:2309.12288. https://arxiv.org/abs/2309.12288
4. **Chen, T., Lin, P., Wang, Z., & Xu, Z.-Q. J. (2025).** *Achilles' Heel of Mamba: Essential difficulties of the Mamba architecture demonstrated by synthetic data.* NeurIPS 2025 / arXiv:2509.17514. https://arxiv.org/abs/2509.17514
5. **Dao, T. (2023/2024).** *FlashAttention-2: Faster Attention with Better Parallelism and Work Partitioning.* ICLR 2024 / arXiv:2307.08691. https://arxiv.org/abs/2307.08691
6. **Dao, T., & Gu, A. (2024).** *Transformers are SSMs: Generalized Models and Efficient Algorithms Through Structured State Space Duality.* ICML 2024 / arXiv:2405.21060. https://arxiv.org/abs/2405.21060
7. **DeepSeek-AI (2024).** *DeepSeek-V3 Technical Report.* arXiv:2412.19437. https://arxiv.org/abs/2412.19437
8. **Dohare, S., Hernandez-Garcia, J. F., Rahman, P., Mahmood, A. R., & Sutton, R. S. (2023/2024).** *Maintaining Plasticity in Deep Continual Learning / Loss of Plasticity in Deep Continual Learning.* Nature 632 / arXiv:2306.13812. https://arxiv.org/abs/2306.13812
9. **Fedus, W., Zoph, B., & Shazeer, N. (2021/2022).** *Switch Transformers: Scaling to Trillion Parameter Models with Simple and Efficient Sparsity.* JMLR 23(120) / arXiv:2101.03961. https://arxiv.org/abs/2101.03961
10. **Fu, X., Hu, Y., Li, B., Feng, Y., Wang, H., Lin, X., Roth, D., Smith, N. A., Ma, W.-C., & Krishna, R. (2024).** *BLINK: Multimodal Large Language Models Can See but Not Perceive.* arXiv:2404.12390. https://arxiv.org/abs/2404.12390
11. **Graves, A. (2016).** *Adaptive Computation Time for Recurrent Neural Networks.* arXiv:1603.08983. https://arxiv.org/abs/1603.08983
12. **Gu, A., & Dao, T. (2023).** *Mamba: Linear-Time Sequence Modeling with Selective State Spaces.* arXiv:2312.00752. https://arxiv.org/abs/2312.00752
13. **Ha, D., & Schmidhuber, J. (2018).** *World Models.* arXiv:1803.10122. https://arxiv.org/abs/1803.10122
14. **Hafner, D., Pasukonis, J., Ba, J., & Lillicrap, T. (2023).** *Mastering Diverse Domains through World Models (DreamerV3).* arXiv:2301.04104. https://arxiv.org/abs/2301.04104
15. **Hoffmann, J., Borgeaud, S., Mensch, A., et al. (2022).** *Training Compute-Optimal Large Language Models (Chinchilla).* arXiv:2203.15556. https://arxiv.org/abs/2203.15556
16. **Hu, R., Cao, Y., Kong, B., Sun, M., & Yuan, K. (2026).** *Synergistic Intra- and Cross-Layer Regularization Losses for MoE Expert Specialization.* arXiv:2602.14159. https://arxiv.org/abs/2602.14159
17. **Huang, R., Wang, Q., Huang, H., Gao, Y., Chen, D., Wang, S., & Wang, W. (2025).** *RepetitionCurse: Measuring and Understanding Router Imbalance in Mixture-of-Experts LLMs under DoS Stress.* arXiv:2512.23995. https://arxiv.org/abs/2512.23995
18. **Kaplan, J., McCandlish, S., Henighan, T., et al. (2020).** *Scaling Laws for Neural Language Models.* arXiv:2001.08361. https://arxiv.org/abs/2001.08361
19. **Kapusuzoglu, B., Pryor, C., Cho, S., Chakraborty, S., Zhang, S.-X., Sahu, S., & Naphade, M. (2026).** *When Load-Balancing Goes Too Far: Expert Pruning in Over-Dispersed Mixture-of-Experts Models.* arXiv:2609.04453. https://arxiv.org/abs/2609.04453
20. **Liu, N. F., Lin, K., Hewitt, J., Paranjape, A., Bevilacqua, M., Petroni, F., & Liang, P. (2023/2024).** *Lost in the Middle: How Language Models Use Long Contexts.* TACL / arXiv:2307.03172. https://arxiv.org/abs/2307.03172
21. **Marcus, G. (2020).** *The Next Decade in AI: Four Steps Towards Robust Artificial Intelligence.* arXiv:2002.06177. https://arxiv.org/abs/2002.06177
22. **Merrill, W., & Sabharwal, A. (2023).** *The Expressive Power of Transformers with Chain of Thought.* ICLR 2024 / arXiv:2310.07923. https://arxiv.org/abs/2310.07923
23. **Shazeer, N., Mirhoseini, A., Maziarz, K., Davis, A., Le, Q., Hinton, G., & Dean, J. (2017).** *Outrageously Large Neural Networks: The Sparsely-Gated Mixture-of-Experts Layer.* ICLR 2017 / arXiv:1701.06538. https://arxiv.org/abs/1701.06538
24. **Tay, Y., Dehghani, M., Tran, V. Q., et al. (2022/2023).** *UL2: Unifying Language Learning Paradigms.* ICLR 2023 / arXiv:2205.05131. https://arxiv.org/abs/2205.05131
25. **Vaswani, A., Shazeer, N., Parmar, N., Uszkoreit, J., Jones, L., Gomez, A. N., Kaiser, L., & Polosukhin, I. (2017).** *Attention Is All You Need.* NeurIPS 2017 / arXiv:1706.03762. https://arxiv.org/abs/1706.03762
26. **Yarden, N., Ben-Artzy, A., & Schwartz, R. (2026).** *Beyond State Capacity: Why Retrieval Breaks in Mamba.* arXiv:2610.19709. https://arxiv.org/abs/2610.19709
27. **Zahn, O., & Chana, S. (2026).** *Selective Memory for Artificial Intelligence: Write-Time Gating with Hierarchical Archiving.* arXiv:2603.15994. https://arxiv.org/abs/2603.15994
