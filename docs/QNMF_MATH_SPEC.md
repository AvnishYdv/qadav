# QNMF Mathematical Specification & Falsifiable Architecture (M9.10)
## Formal Algebraic, Dynamical, and Information-Theoretic Foundations for the Qadav Neuro-Causal Memory Fabric

**Document Version:** 1.0.0  
**Milestone:** Qadav Phase 10 — M9.10 Mathematical Specification  
**Author:** AI Research Scientist, Mathematician & Compiler Architecture Team  
**Project:** Qadav (`.ydv`) — Autonomous AGI Programming Language  
**Target Repository:** `c:\Users\avnis\Downloads\Qadav`  
**Classification:** Foundational Theoretical & Mathematical Specification (Zero Code, Research Spec Only)  
**Prerequisite:** [`docs/QNMF_PRIOR_ART_AUDIT.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_PRIOR_ART_AUDIT.md) (Accepted & Integrated)  

---

## Executive Overview & Mathematical Notation

The **Qadav Neuro-Causal Memory Fabric (QNMF)** is an explicit non-Transformer cognitive runtime architecture designed to resolve the foundational mathematical and architectural failure modes documented in the M9.9 Prior-Art Audit ([`docs/QNMF_PRIOR_ART_AUDIT.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_PRIOR_ART_AUDIT.md)): quadratic context scaling ($O(N^2)$), circuit depth constraints ($TC^0$), the reversal curse ($P(B|A) \not\Rightarrow P(A|B)$), delayed relevance gating in State-Space Models (SSMs), the inability of continuous neural memory to perform surgical unlearning, and routing collapse under adversarial repetition in Mixture-of-Experts (MoE).

This document serves as the formal mathematical specification for QNMF. It details the state spaces, continuous-to-discrete dynamical equations, multi-timescale memory operators, regularized optimal-transport routing, predictive world rollouts, and evidence-verification predicates. Every theoretical proposition formulated herein is paired with an explicit null hypothesis ($H_0$), alternative hypothesis ($H_1$), and a mathematically falsifiable empirical protocol.

### Mathematical Notation Reference

| Symbol | Mathematical Domain | Description |
|---|---|---|
| $t \in \mathbb{N}$ | Discrete sequence index | Discrete token or cognitive execution step |
| $x_t \in \mathbb{R}^{d_{\text{in}}}$ | Input vector space | Embedded token or sensory observation at step $t$ |
| $S_t \in \mathbb{R}^{d_s}$ | Continuous dynamical state | Causal recurrent core state at step $t$ |
| $M_s^{(t)} \in \mathbb{R}^{N_s \times d_m}$ | Matrix manifold | Short-term sliding-window sensory buffer |
| $M_w^{(t)} \in \mathbb{R}^{N_w \times d_m}$ | Matrix manifold | Working associative test-time memory matrix |
| $M_l^{(t)} \in \mathbb{R}^{N_l \times d_m}$ | Keyed tensor manifold | Long-term crystalline topological memory store |
| $r_t \in \Delta^{K-1}$ | Probability simplex | Sparse cognitive routing vector over $K$ expert fabrics |
| $z_t \in \mathbb{R}^{d_z}$ | Latent representation | World-model latent state representation |
| $\kappa_t \in \mathcal{K}$ | Cryptographic certificate space | Immutable proof witness tuple $(\text{id}, h_{\text{pre}}, h_{\text{op}}, h_{\text{post}}, w, \sigma)$ |
| $B_t \in \mathbb{R}_{\ge 0}$ | Scalar budget | Adaptive computational ponder budget |
| $\odot$ | Hadamard product | Element-wise vector/matrix multiplication |
| $\otimes$ | Outer product / Kronecker | Tensor product of vector spaces |
| $\|\cdot\|_2, \|\cdot\|_F$ | Vector/Matrix norms | Euclidean vector norm and Frobenius matrix norm |
| $\mathcal{W}_{\varepsilon}(u, v)$ | Optimal transport distance | Entropic regularized Wasserstein distance |

---

## Part A — Mathematical Formalization of the 6 QNMF Pillars

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                              QNMF UNIFIED COGNITIVE RUNTIME                            │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  P1: Causal State Dynamics          ──► S_{t+1} = Λ_t ⊙ S_t + Φ(x_t) + Ψ(M_w^{(t)})    │
│  P2: Multi-Timescale Memory        ──► M_s (Fast) ──► M_w (Plastic) ──► M_l (Crystalline)│
│  P3: Adaptive Cognitive Routing     ──► r_t = Sinkhorn_ε(C(x_t), u, v) ∈ Δ^{K-1}       │
│  P4: Predictive World Model         ──► z_{t+k} = T_ω^{(k)}(z_t, a_{t:t+k-1})          │
│  P5: Evidence-Guided Reasoning      ──► V(κ_t, h_{premise}, h_{concl}) ∈ {ACC, REF, DEC}│
│  P6: Adaptive Compute Controller    ──► C_ψ(S_t, Uncertainty_t) ──► B_t, Stop(S_t)     │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

### Pillar 1: Causal State Dynamics (P1)

#### 1.1 State Definition
The core state vector is denoted by $S_t \in \mathbb{R}^{d_s}$, initialized at $S_0 = \mathbf{0}$. Unlike pure autoregressive attention, $S_t$ represents a compressed, continuous-time causal state of the cognitive stream.

#### 1.2 Continuous-to-Discrete Transition Formulation
The underlying continuous-time dynamics are governed by a bilaterally symmetric linear differential operator:
$$\frac{d h(t)}{dt} = A(t) h(t) + B(t) u(t)$$
where $A(t) \in \mathbb{R}^{d_s \times d_s}$ is parameterized as a diagonal decay matrix to ensure computational tractability, and $B(t) \in \mathbb{R}^{d_s \times d_{\text{in}}}$ is an input projection operator.

To eliminate the **Asymmetry Bias** documented in M9.9 Audit Claim B.1 ([`docs/QNMF_PRIOR_ART_AUDIT.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_PRIOR_ART_AUDIT.md)), QNMF removes the asymmetric 1D convolution of Mamba and replaces it with a **Commutative Bilateral Encoding Tensor** $\Phi(x_t) \in \mathbb{R}^{d_s}$:
$$\Phi(x_t) = \frac{1}{2} \left( W_u x_t + W_v \mathcal{T}(x_t) \right) + \text{Sym}\left(W_a x_t \otimes W_b x_t\right) v_{\text{proj}}$$
where $\mathcal{T}(\cdot)$ is an involution operator ($\mathcal{T}(\mathcal{T}(x)) = x$), and $\text{Sym}(M) = \frac{1}{2}(M + M^\top)$ enforces algebraic symmetry over relational pairs.

Discretizing the continuous operator over step size $\Delta_t = \text{softplus}(w_\Delta^\top x_t + b_\Delta)$ yields the discrete transition equation:
$$\Lambda_t = \exp\left(-\Delta_t \cdot \text{diag}(\alpha_1, \alpha_2, \dots, \alpha_{d_s})\right) \in (0, 1)^{d_s}$$
$$S_{t+1} = f_\theta(S_t, x_t, M_w^{(t)}, r_t) = \Lambda_t \odot S_t + (1 - \Lambda_t) \odot \Phi(x_t) + W_{\text{mem}} \text{Read}(M_w^{(t)}, S_t) + \sum_{k=1}^K r_{t, k} E_k(S_t)$$

#### 1.3 Learning Objective
The parameters $\theta = \{W_u, W_v, W_a, W_b, w_\Delta, \alpha, W_{\text{mem}}\}$ are optimized via sequence modeling loss combined with a spectral radius stability regularizer:
$$\mathcal{L}_{\text{P1}}(\theta) = \mathbb{E}_{(x, y) \sim \mathcal{D}} \left[ -\sum_{t=1}^T \log P(y_t \mid S_t) \right] + \lambda_{\text{spec}} \mathcal{R}_{\text{spectral}}(\Lambda)$$
where the spectral regularizer strictly confines state eigenvalues within the open unit disk:
$$\mathcal{R}_{\text{spectral}}(\Lambda) = \frac{1}{d_s} \sum_{i=1}^{d_s} \max\left(0, \Lambda_{t, i} - (1 - \epsilon)\right)^2, \quad \epsilon = 10^{-4}$$

#### 1.4 Complexity Analysis
* **Time Complexity:** Matrix-vector operations scale as $O(d_s \cdot d_{\text{in}} + d_s)$ per token step. The full sequence of length $L$ evaluates in $O(L \cdot d_s)$ via associative parallel prefix scans.
* **Space Complexity:** $O(d_s)$ activation memory per step during inference.
* **Gradient Derivation:** Gradients $\frac{\partial \mathcal{L}}{\partial S_t}$ backpropagate through time via the recurrence:
  $$\frac{\partial \mathcal{L}}{\partial S_t} = \frac{\partial \ell_t}{\partial S_t} + \Lambda_{t+1} \odot \frac{\partial \mathcal{L}}{\partial S_{t+1}}$$
  Because $\Lambda_{t+1} \in (0, 1)^{d_s}$, gradients decay exponentially with distance unless refreshed by the memory coupling term $W_{\text{mem}}$, preventing exploding gradients.

#### 1.5 Falsifiable Prediction (P1)
* **Prediction P1.1:** On synthetic inverted sequence prediction (mirror sequences of length $L \in [64, 512]$), QNMF will achieve $\ge 98.0\%$ sequence accuracy, whereas Mamba-1/2 will achieve $< 45.0\%$ accuracy due to its asymmetric convolution bottleneck (as proven in NeurIPS 2025).

---

### Pillar 2: Multi-Timescale Neural Memory (P2)

```
MULTI-TIMESCALE MEMORY FLOW:
┌─────────────────────────┐
│ Input Token x_t, S_t    │
└────────────┬────────────┘
             ▼
┌─────────────────────────┐
│ Short-Term Buffer M_s   │ ──► Ring-buffer FIFO (Recent k_s tokens)
└────────────┬────────────┘
             │ Surprise Gating: σ_t = ||∇_{M_w} L||_2 > τ_w
             ▼
┌─────────────────────────┐
│ Working Memory M_w      │ ──► Test-time plastic updates: M_w^{(t+1)} = (1-η)M_w + η (k ⊗ v)
└────────────┬────────────┘
             │ Crystallization Gating: Frequency > τ_c
             ▼
┌─────────────────────────┐
│ Crystalline Store M_l   │ ──► Cryptographically keyed, O(1) revocable slots
└─────────────────────────┘
```

#### 2.1 State Definitions
QNMF partitions associative memory across three distinct temporal scales to eliminate catastrophic forgetting and provide surgical unlearning capabilities:
1. **Short-Term Memory ($M_s$):** A continuous ring-buffer of dimension $N_s \times d_m$ caching the most recent $N_s$ hidden representations.
2. **Working Associative Memory ($M_w$):** A plastic matrix manifold $M_w \in \mathbb{R}^{N_w \times d_m}$ supporting test-time online parameter updates.
3. **Long-Term Crystalline Store ($M_l$):** An indexed key-value repository $(K_l, V_l)$ where $K_l \in \mathbb{R}^{N_l \times d_k}$ and $V_l \in \mathbb{R}^{N_l \times d_v}$, partitioned into explicit addressable slots.

#### 2.2 Transition & Transfer Equations
Information cascades through memory based on mathematical **Surprise** and **Utility Metrics**:

1. *Surprise Metric Calculation:*
   $$\sigma_t = \left\| x_t - \text{Predict}(S_t) \right\|_2^2$$
2. *Short-to-Working Transfer ($M_s \to M_w$):*
   If $\sigma_t > \tau_{\text{surprise}}$, key-value binding $k_t = W_k x_t$, $v_t = W_v x_t$ is injected into working memory via a bounded outer-product update:
   $$M_w^{(t+1)} = (1 - \gamma_t) M_w^{(t)} + \eta_t \cdot \frac{k_t \otimes (v_t - M_w^{(t)} k_t)}{1 + \|k_t\|_2^2}$$
   where $\gamma_t = \text{sigmoid}(w_\gamma^\top S_t)$ is an adaptive forgetting rate, and $\eta_t = \text{sigmoid}(w_\eta^\top S_t)$ is the write learning rate.
3. *Working-to-Crystalline Crystallization ($M_w \to M_l$):*
   When an associative pattern $(k_i, v_i)$ achieves cumulative retrieval frequency $F(k_i) > \tau_{\text{cryst}}$, it is normalized and committed to an empty or least-recently-used slot $j \in \{1, \dots, N_l\}$ in $M_l$:
   $$K_{l, j} = \frac{k_i}{\|k_i\|_2}, \quad V_{l, j} = v_i, \quad \text{Status}_j = \text{ACTIVE}$$
4. *Surgical Revocation Operator ($O(1)$ Deletion):*
   Unlike continuous neural memory (Titans), any targeted slot $j$ can be surgically wiped in $O(1)$ time by key hash invalidation without retraining:
   $$\text{Revoke}(h_k) \implies K_{l, j} \leftarrow \mathbf{0}, \quad V_{l, j} \leftarrow \mathbf{0}, \quad \text{Status}_j \leftarrow \text{REVOKED}$$

#### 2.3 Learning Objective
$$\mathcal{L}_{\text{P2}} = \mathcal{L}_{\text{retrieve}}(M_s, M_w, M_l) + \lambda_{\text{ortho}} \left\| K_l K_l^\top - I_{N_l} \right\|_F^2 + \lambda_{\text{sparse}} \sum_{t=1}^T \eta_t$$

#### 2.4 Complexity Analysis
* **Time Complexity:** Querying $M_w$ is $O(d_m^2)$; querying $M_l$ via approximate nearest neighbor (ANN) or clustered projections is $O(\log N_l \cdot d_k)$. Total query time is sub-linear in total stored facts.
* **Space Complexity:** $O(N_s d_m + N_w d_m + N_l(d_k + d_v))$ — strictly linear in capacity, independent of sequence context length $L$.

#### 2.5 Falsifiable Prediction (P2)
* **Prediction P2.1:** Upon executing a single surgical deletion instruction $\text{Revoke}(h_{\text{target}})$, the recall probability $P(\text{target} \mid \text{query})$ drops to exactly $0.00\%$ within $O(1)$ operations, while retrieval accuracy on adjacent retained memories drops by $< 0.05\%$. Under identical test conditions, Titans (Behrouz et al., 2025) cannot surgically delete the target fact without gradient unlearning that reduces adjacent accuracy by $\ge 12.0\%$.

---

### Pillar 3: Adaptive Sparse Cognitive Routing (P3)

#### 3.1 State Definition
Let $K$ denote the number of specialized cognitive expert fabrics (e.g., Symbolic Deduction, Numerical Calculus, Multimodal Parsing, Memory Consolidation). The routing assignment vector at step $t$ is represented by $r_t \in \Delta^{K-1} = \{r \in \mathbb{R}^K \mid \sum_{k=1}^K r_k = 1, r_k \ge 0\}$.

#### 3.2 Regularized Optimal Transport Transition Formulation
To eliminate the **RepetitionCurse DoS Vulnerability** and **Pseudo-Balancing Flaws** documented in M9.9 Audit Claims D.1, D.2, and D.4 ([`docs/QNMF_PRIOR_ART_AUDIT.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_PRIOR_ART_AUDIT.md)), QNMF replaces heuristic auxiliary softmax balancing with a **Batch-Constrained Entropic Optimal Transport (Sinkhorn) Formulation**:

Given a token batch of size $B$ and expert capacity target $c = \frac{B}{K} \mathbf{1}_K$, let $C \in \mathbb{R}^{B \times K}$ represent the cognitive dispatch cost matrix:
$$C_{i, k} = \frac{1}{2} \left\| W_{\text{gate}} x_i - \mu_k \right\|_2^2$$
where $\mu_k \in \mathbb{R}^{d_{\text{gate}}}$ represents the centroid vector of expert fabric $k$.

The optimal transportation plan $P^* \in \mathbb{R}^{B \times K}$ solves:
$$\min_{P \in \Pi(u, v)} \sum_{i=1}^B \sum_{k=1}^K P_{i, k} C_{i, k} - \varepsilon \mathcal{H}(P)$$
where $\Pi(u, v) = \{P \in \mathbb{R}_+^{B \times K} \mid P \mathbf{1}_K = \frac{1}{B} \mathbf{1}_B, P^\top \mathbf{1}_B = \frac{1}{K} \mathbf{1}_K\}$, and $\mathcal{H}(P) = -\sum_{i, k} P_{i, k} (\log P_{i, k} - 1)$ is the Shannon entropy regularizer.

The unique optimal transport solution is computed via the iterative Sinkhorn-Knopp algorithm:
$$P^* = \text{diag}(u^{(\tau)}) \exp\left(-\frac{C}{\varepsilon}\right) \text{diag}(v^{(\tau)})$$
with coordinate scaling updates:
$$u^{(\tau+1)} = \frac{\frac{1}{B} \mathbf{1}_B}{\exp(-C/\varepsilon) v^{(\tau)}}, \quad v^{(\tau+1)} = \frac{\frac{1}{K} \mathbf{1}_K}{\exp(-C/\varepsilon)^\top u^{(\tau+1)}}$$
After convergence (typically $\tau \le 5$ iterations), the sparse routing vector for token $i$ is:
$$r_i = \text{TopK}\left(P^*_{i, :}, k_{\text{active}}\right), \quad k_{\text{active}} \ll K$$

#### 3.3 Learning Objective
$$\mathcal{L}_{\text{P3}}(\phi) = \mathcal{L}_{\text{task}} + \lambda_{\text{OT}} \mathcal{W}_{\varepsilon}(u, v) + \lambda_{\text{disp}} \sum_{j \neq k} \max\left(0, \cos(\mu_j, \mu_k) - \delta_{\text{sep}}\right)^2$$
where $\mathcal{W}_{\varepsilon}(u, v) = \langle P^*, C \rangle - \varepsilon \mathcal{H}(P^*)$ is the Sinkhorn divergence, which is provably convex and differentiable with respect to gate parameters $W_{\text{gate}}$.

#### 3.4 Complexity Analysis
* **Time Complexity:** Computing cost matrix $C$ is $O(B \cdot K \cdot d_{\text{gate}})$. Each Sinkhorn iteration is $O(B \cdot K)$. For $\tau = 5$, routing overhead is $< 3\%$ of total layer computation.
* **Space Complexity:** $O(B \cdot K)$ memory for transport matrices.

#### 3.5 Falsifiable Prediction (P3)
* **Prediction P3.1:** Under the RepetitionCurse benchmark (input sequence consisting of identical token phrase repeated $10^4$ times), the maximum expert load variance across all $K$ experts will satisfy $\text{Var}(\text{Load}) < 0.02$, and Time-to-First-Token (TTFT) latency will degrade by $< 4.0\%$. Under identical test conditions, standard MoE routing (Switch Transformer / Mixtral heuristic softmax) will route $> 95.0\%$ of tokens to a single expert, producing $\text{Var}(\text{Load}) > 0.45$ and a latency spike exceeding $400\%$.

---

### Pillar 4: Predictive Latent World Model (P4)

#### 4.1 State Definition
Let $z_t = g_\omega(S_t) \in \mathbb{R}^{d_z}$ be a low-dimensional causal latent representation of the environment state. Let $a_t \in \mathcal{A}$ represent a discrete or continuous cognitive action (e.g., query memory, expand theorem, execute tool, invoke deductive rule).

#### 4.2 Latent Rollout Dynamics
Rather than autoregressively rolling out raw linguistic tokens in token space (which compounds sampling variance and hallucinations), the QNMF World Model performs transition rollouts entirely in the **Causal Latent Geometry**:
$$\hat{z}_{t+1} = T_\omega(z_t, a_t) = z_t + \Delta \tau \cdot \mathcal{F}_\omega(z_t, a_t)$$
where $\mathcal{F}_\omega$ is a residual neural vector field. Multi-step recursive rollout up to lookahead horizon $H \in \mathbb{N}$ is defined by:
$$\hat{z}_{t+k} = T_\omega\left(\hat{z}_{t+k-1}, a_{t+k-1}\right), \quad k \in \{1, 2, \dots, H\}$$

#### 4.3 Objective Function (Non-Contrastive Latent Consistency)
To prevent representation collapse without requiring expensive negative sample pairs, QNMF employs a Variance-Invariance-Covariance Regularized (VICReg-style) transition loss:
$$\mathcal{L}_{\text{P4}}(\omega) = \sum_{k=1}^H \gamma^k \left\| z_{t+k} - \hat{z}_{t+k} \right\|_2^2 + \lambda_{\text{var}} \mathcal{V}(Z) + \lambda_{\text{cov}} \mathcal{C}(Z)$$
where:
$$\mathcal{V}(Z) = \frac{1}{d_z} \sum_{j=1}^{d_z} \max\left(0, 1 - \sqrt{\text{Var}(z_{:, j}) + \epsilon}\right)$$
$$\mathcal{C}(Z) = \frac{1}{d_z} \sum_{i \neq j} \left( \text{Cov}(Z)_{i, j} \right)^2$$

#### 4.4 Complexity Analysis
* **Time Complexity:** $O(H \cdot d_z^2)$ FLOPs per mental simulation. Because $d_z \ll d_s$ (typically $d_z = 256$ vs. $d_s = 4096$), a depth-$H=10$ latent rollout costs less compute than generating a single autoregressive output token.
* **Space Complexity:** $O(H \cdot d_z)$ memory for trajectory caching.

#### 4.5 Falsifiable Prediction (P4)
* **Prediction P4.1:** In goal-directed multi-hop planning tasks (BlocksWorld, ProofGraph with diameter $D=8$), QNMF latent rollouts will achieve $\ge 82.0\%$ plan success within $B=16$ mental simulations, whereas autoregressive token-level Chain-of-Thought search will achieve $< 40.0\%$ success due to compounding branch hallucinations.

---

### Pillar 5: Evidence-Guided Reasoning (P5)

#### 5.1 Formal Certificate Definition
Every deductive, inductive, or tool-assisted inference step must emit a cryptographic proof certificate $\kappa_t \in \mathcal{K}$:
$$\kappa_t = \Big\langle \text{step\_id} \in \mathbb{N}, \; h_{\text{premise}} \in \{0, 1\}^{256}, \; \text{op\_code} \in \mathcal{O}, \; h_{\text{conclusion}} \in \{0, 1\}^{256}, \; w_t \in \mathcal{W}, \; c_t \in [0, 1] \Big\rangle$$
where:
* $h_{\text{premise}} = \text{SHA256}(\text{Premise Text} \mid S_t)$
* $\text{op\_code} \in \{\text{DEDUCE}, \text{INDUCE}, \text{ABDUCE}, \text{VERIFY}, \text{TOOL\_CALL}\}$
* $w_t$ is the explicit proof witness (axiom pointer, guideline hash, or formal ground truth hash)
* $c_t \in [0, 1]$ is the mathematically calibrated Bayesian confidence score:
  $$c_t = \frac{1}{1 + \exp\left(-\beta_0 - \beta_1 \cdot \text{EvidenceMargin}(S_t)\right)}$$

#### 5.2 Deterministic Verification Function
The verification operator $V: \mathcal{K} \times \{0, 1\}^{256} \times \{0, 1\}^{256} \to \{\text{ACCEPT}, \text{REFUSE}, \text{DECOHERE}\}$ is defined as:
$$V(\kappa_t, h_{\text{premise}}, h_{\text{conclusion}}) = \begin{cases} 
\text{ACCEPT} & \text{if } \text{SHA256}(h_{\text{premise}} \parallel \text{op\_code} \parallel w_t) = h_{\text{conclusion}} \text{ and } c_t \ge \tau_{\text{safe}} \\
\text{REFUSE} & \text{if } c_t < \tau_{\text{safe}} \text{ (Calibrated Safe Abstention)} \\
\text{DECOHERE} & \text{if } \text{SHA256}(h_{\text{premise}} \parallel \text{op\_code} \parallel w_t) \neq h_{\text{conclusion}} \text{ (Tamper Detected)}
\end{cases}$$
When $V(\cdot) = \text{DECOHERE}$, the execution engine invokes hardware-inspired memory zeroization via `qadav_qeeh_trigger_decoherence` and halts the process via Windows API `ExitProcess(0xDEADBEEF)`.

#### 5.3 Objective Function
$$\mathcal{L}_{\text{P5}} = \mathcal{L}_{\text{task}} + \lambda_{\text{calib}} \text{ECE}(c, \mathbf{1}_{\text{correct}}) + \lambda_{\text{abstain}} \max(0, \tau_{\text{safe}} - c_t) \cdot \mathbb{I}_{\text{incorrect}}$$
where $\text{ECE}$ is the Expected Calibration Error, enforcing that a confidence $c_t = 0.90$ corresponds to an empirical error rate of exactly $10.0\%$.

#### 5.4 Complexity Analysis
* **Time Complexity:** $O(1)$ constant-time cryptographic SHA-256 evaluation per step ($\sim 1.2$ microseconds on modern CPU silicon).
* **Space Complexity:** 128 bytes per proof certificate record.

#### 5.5 Falsifiable Prediction (P5)
* **Prediction P5.1:** When subjected to out-of-distribution ungrounded queries (where correct answer cannot be derived from context), QNMF will execute calibrated abstention ($\text{REFUSE}$) in $\ge 96.0\%$ of cases with zero hallucinated false positive deductions ($c_t < \tau_{\text{safe}}$), whereas uncalibrated dense Transformers hallucinate unsupported claims with high confidence ($c > 0.85$) in $\ge 68.0\%$ of cases.

---

### Pillar 6: Adaptive Compute Controller / ECRA (P6)

#### 6.1 State Definition
Let $B_t \in \mathbb{R}_{\ge 0}$ represent the dynamic computational budget (measured in internal deliberation cycles) assigned to token or task $t$. The controller is parameterized by a lightweight neural recurrent unit:
$$h_{\text{ctrl}}^{(t)} = \text{GRU}_\psi\left(x_t, S_t, \text{Uncertainty}_t, h_{\text{ctrl}}^{(t-1)}\right)$$

#### 6.2 Compute Allocation & Halting Mechanics
The total deliberation budget is computed via a bounded monotonic mapping:
$$B_t = B_{\text{min}} + (B_{\text{max}} - B_{\text{min}}) \cdot \text{sigmoid}\left(w_B^\top h_{\text{ctrl}}^{(t)} + b_B\right)$$
During multi-step internal reasoning ($s = 1, 2, \dots, \lfloor B_t \rfloor$), the halting operator evaluates:
$$p_{\text{halt}}^{(s)} = \text{sigmoid}\left(w_{\text{halt}}^\top S_t^{(s)} + b_{\text{halt}}\right)$$
$$\text{Stop}(s) = \begin{cases} 
\text{TRUE} & \text{if } \sum_{i=1}^s p_{\text{halt}}^{(i)} \ge 1.0 - \epsilon_{\text{halt}} \quad \text{or} \quad s \ge \lfloor B_t \rfloor \\
\text{FALSE} & \text{otherwise}
\end{cases}$$

#### 6.3 Learning Objective (Ponder Cost Regularization)
Following the Adaptive Computation Time (ACT) formulation:
$$\mathcal{L}_{\text{P6}}(\psi) = \mathcal{L}_{\text{task}} + \lambda_{\text{ponder}} \sum_{t=1}^T \left( N_t + R_t \right)$$
where $N_t = \min\{s \mid \sum_{i=1}^s p_{\text{halt}}^{(i)} \ge 1 - \epsilon\}$, and $R_t = 1 - \sum_{i=1}^{N_t - 1} p_{\text{halt}}^{(i)}$ is the residual halting probability.

#### 6.4 Complexity Analysis
* **Time Complexity:** Average runtime per token is $O(\bar{N} \cdot d_s)$, where $\bar{N} \in [B_{\text{min}}, B_{\text{max}}]$ dynamically shrinks to $B_{\text{min}} = 1$ for trivial syntactic tokens and expands to $B_{\text{max}}$ for complex multi-step deductions.
* **Space Complexity:** $O(d_{\text{ctrl}})$.

#### 6.5 Falsifiable Prediction (P6)
* **Prediction P6.1:** Across a heterogeneous test benchmark comprising 50% trivial punctuation/grammar tokens and 50% complex arithmetic reasoning tokens, QNMF will allocate an average of $1.1$ deliberation steps to trivial tokens and $6.8$ steps to arithmetic tokens, reducing total inference FLOPs by $\ge 42.0\%$ compared to fixed-depth Transformer baselines while matching or exceeding task accuracy.

---

## Part B — Information-Theoretic Propositions & Theoretical Arguments

```
PROPOSITION PROOF ARCHITECTURE:
┌────────────────────────────────────────────────────────────────────────┐
│ Proposition 1 (Context Scaling): QNMF O(L·N_m) vs Transformer O(L²)   │
├────────────────────────────────────────────────────────────────────────┤
│ Proposition 2 (Length Invariance): S_{t+1} = f_θ(S_t, x_t) ∉ TC⁰       │
├────────────────────────────────────────────────────────────────────────┤
│ Proposition 3 (Routing Stability): Sinkhorn W_ε(u,v) Provable Convexity│
├────────────────────────────────────────────────────────────────────────┤
│ Proposition 4 (Plasticity Separation): θ_{slow} ⊥ M_w^{(t)} Ortho SVD │
└────────────────────────────────────────────────────────────────────────┘
```

### Proposition 1: Sub-Quadratic Context Scaling Efficiency
* **Statement (Hypothesis):** At context sequence length $L$, the dense multi-head self-attention operator requires $O(L^2 \cdot d)$ floating-point operations and $O(L \cdot d)$ cumulative KV-cache storage per sequence. The QNMF runtime, operating with a bounded multi-timescale memory capacity $N_m = N_s + N_w + N_l$, evaluates sequence updates in $O(L \cdot (d_s + N_m \cdot d_m))$ FLOPs. As context length $L \to \infty$, QNMF achieves an asymptotic computational speedup of $O(L / N_m)$ per token over full self-attention while maintaining a strictly bounded memory footprint $O(N_m)$.
* **Formal Derivation Sketch:**
  Let sequence length be $L$. In full self-attention:
  $$\text{FLOPs}_{\text{Attn}} = 4 L^2 d + 2 L^2 = \Omega(L^2)$$
  $$\text{Memory}_{\text{Attn}} = 2 L \cdot n_{\text{layers}} \cdot d_{\text{kv}} = \Omega(L)$$
  In QNMF, each step $t \in \{1, \dots, L\}$ computes a state update $S_{t+1} = f_\theta(S_t, x_t)$ and queries memory $M^{(t)}$:
  $$\text{FLOPs}_{\text{QNMF}} = L \cdot \left( c_1 d_s^2 + c_2 N_m d_m \right) = O(L)$$
  $$\text{Memory}_{\text{QNMF}} = c_3 d_s + c_4 N_m d_m = O(1) \quad (\text{independent of } L)$$
  Taking the ratio of per-token FLOPs:
  $$\lim_{L \to \infty} \frac{\text{Cost}_{\text{Attn}} / L}{\text{Cost}_{\text{QNMF}} / L} = \lim_{L \to \infty} \frac{O(L)}{O(N_m)} = \infty$$
* **Null Hypothesis ($H_0$):** At context lengths $L \in \{1\text{K}, 10\text{K}, 100\text{K}, 1\text{M}\}$, QNMF wall-clock inference execution latency per token scales identically to or worse than FlashAttention-2 Transformer baselines ($t_{\text{token}}(L) \propto L$).
* **Alternative Hypothesis ($H_1$):** QNMF wall-clock execution latency per token remains flat ($O(1)$) across all $L \in [1\text{K}, 1\text{M}]$, demonstrating an asymptotic speedup $> 10\times$ at $L = 100\text{K}$ and $> 80\times$ at $L = 1\text{M}$ over FlashAttention-2.
* **Acceptance Threshold:** Wall-clock per-token generation latency must vary by $< 8.0\%$ between $L=1\text{K}$ and $L=1\text{M}$ on identical CPU/GPU hardware.
* **Specific Baseline:** LLaMA-3-8B / FlashAttention-2 running on 1M context.

---

### Proposition 2: Length-Invariant Algorithmic Generalization Outside $TC^0$
* **Statement (Hypothesis):** A standard constant-depth Transformer without dynamic scratchpads cannot compute discrete state transitions outside the circuit complexity class uniform $TC^0$, failing to generalize algorithmic tasks (e.g., modular parity, graph cycle detection, balanced parentheses) to sequence lengths $L' \gg L_{\text{train}}$. QNMF's recurrent causal state $S_{t+1} = f_\theta(S_t, x_t)$, augmented with discrete modular certificate predicates (Pillar 5), constitutes a dynamical finite-state transducer with pushdown capabilities, achieving length-invariant generalization for arbitrary $L' > L_{\text{train}}$.
* **Formal Derivation Sketch:**
  From Yuan et al. (EACL 2026, long.87), a Transformer with $D$ layers and log-precision activations computes functions in $TC^0$. The language of parity $\text{PARITY} = \{x \in \{0, 1\}^* \mid \sum x_i \equiv 1 \pmod 2\}$ is provably not in $AC^0$ and requires depth $\Omega(L / \log L)$ for constant-width threshold circuits. In QNMF:
  $$S_{t+1} = S_t \oplus x_t \quad (\text{represented in the discrete certificate manifold})$$
  Because the recurrence operator $f_\theta$ is applied sequentially $L'$ times without altering parameter weights or positional basis rotations, the transition function is strictly length-invariant:
  $$f_\theta(S_{t}, x_{t}) = f_\theta(S_{t+k}, x_{t+k}) \quad \forall k$$
* **Null Hypothesis ($H_0$):** When trained on arithmetic parity sequences of length $L \le 64$, QNMF accuracy on sequences of length $L' \in [128, 4096]$ degrades at the same rate as standard Transformers (dropping to random chance $50.0\%$).
* **Alternative Hypothesis ($H_1$):** QNMF maintains $\ge 99.5\%$ exact accuracy on parity sequences up to $L' = 4096$ without parameter retraining.
* **Acceptance Threshold:** Zero accuracy degradation ($\Delta < 0.5\%$) on $L' = 64 \times L_{\text{train}}$.
* **Specific Baseline:** GPT-4o-mini and Transformer-XL (trained at $L=64$).

---

### Proposition 3: Optimal Transport Routing Invariance under Adversarial Repetition
* **Statement (Hypothesis):** Softmax-based MoE gating functions $G(x) = \text{Softmax}(\text{TopK}(W_g x, k))$ equipped with heuristic auxiliary balancing losses $\mathcal{L}_{\text{aux}} = \alpha E \sum f_i P_i$ lack convexity and fail under low-entropy out-of-distribution prompts, causing all tokens to concentrate onto a single expert. In contrast, QNMF's entropic Optimal Transport routing (Pillar 3) defines a strictly convex optimization problem over the transportation polytope $\Pi(u, v)$ with guaranteed convergence to uniform marginal expert distributions, provably eliminating the RepetitionCurse DoS vulnerability.
* **Formal Derivation Sketch:**
  Let the objective be $f(P) = \langle P, C \rangle - \varepsilon \mathcal{H}(P)$. The negative entropy $\mathcal{H}(P)$ is strictly concave, hence $- \varepsilon \mathcal{H}(P)$ is strictly convex with modulus $\varepsilon$. The feasible set $\Pi(u, v)$ is the intersection of affine hyperplanes, which is compact and convex. By the Bregman projection theorem, the Sinkhorn-Knopp algorithm converges at a geometric rate:
  $$\|u^{(\tau)} - u^*\| \le O(\rho^{2\tau}), \quad \rho = \frac{1 - \sqrt{\kappa}}{1 + \sqrt{\kappa}} < 1$$
  where $\kappa = \min_{i, j} \exp(-C_{i, j}/\varepsilon) > 0$. The marginal constraint $\sum_{i=1}^B P_{i, k} = \frac{B}{K}$ guarantees that expert $k$ receives exactly its uniform quota of tokens across the batch, regardless of the cost matrix $C$.
* **Null Hypothesis ($H_0$):** Under adversarial token repetition ($10^4$ identical tokens), QNMF routes $> 80\%$ of tokens to a single expert, exhibiting the same load collapse as Switch Transformer.
* **Alternative Hypothesis ($H_1$):** Under adversarial repetition, QNMF distributes tokens evenly across all experts such that every expert receives $\frac{B}{K} \pm 5\%$ of total tokens, preserving bounded execution latency.
* **Acceptance Threshold:** Gini coefficient of expert load $G_{\text{load}} \le 0.05$ under adversarial repetition.
* **Specific Baseline:** Mixtral-8x7B / Switch Transformer MoE.

---

### Proposition 4: Dual-Manifold Plasticity Separation against Catastrophic Forgetting
* **Statement (Hypothesis):** Catastrophic forgetting and representation collapse in deep learning arise from monolithic parameter updates where gradient descent modifies shared feature weights along non-orthogonal task vectors. By segregating the cognitive fabric into an immutable crystalline parameter manifold $\theta_{\text{slow}}$ (slow learning rate, high regularization) and a dynamic, slotted associative memory $M_l$ (fast discrete slot updates), QNMF achieves zero forgetting on preserved factual slots ($0.0\%$ degradation) over sequential task curricula without raw data experience replay.
* **Formal Derivation Sketch:**
  Let total capability be represented by the decomposition $y = g(x; \theta_{\text{slow}}) + \text{Query}(x, M_l)$. When training sequentially on Task 1 then Task 2:
  $$\Delta \theta_{\text{slow}} \le \eta_{\text{slow}} \cdot \text{Proj}_{\text{null}(\nabla \mathcal{L}_{T_1})}(\nabla \mathcal{L}_{T_2})$$
  $$\Delta M_l = \text{WriteSlot}\left(\text{Slot}_{\text{empty}}, \; \text{Hash}(x_{T_2}), \; v_{T_2}\right)$$
  Because new factual knowledge is stored in previously unoccupied orthogonal memory slots $\text{Slot}_j$ ($j \neq i$), the memory read operator on Task 1 queries satisfies:
  $$\text{Query}\left(x_{T_1}, M_l^{(T_2)}\right) = \text{Query}\left(x_{T_1}, M_l^{(T_1)}\right) \pm 0$$
* **Null Hypothesis ($H_0$):** On a curriculum of 50 sequential disjoint classification/QA tasks, QNMF retention accuracy on Task 1 after completing Task 50 degrades by $\ge 40.0\%$ (matching standard fine-tuned LLMs).
* **Alternative Hypothesis ($H_1$):** QNMF retention accuracy on Task 1 degrades by $< 3.0\%$ without experience replay.
* **Acceptance Threshold:** Backward Transfer $\text{BWT} \ge -0.03$.
* **Specific Baseline:** LLaMA-3-8B sequential LoRA fine-tuning vs. standard AdamW.

---

## Part C — Computational Complexity & Scaling Analysis

### Computational Complexity Profile

| Component | Operation Phase | Time Complexity (Per Token) | Space Complexity (Activations) | Hardware Bound |
|---|---|---|---|---|
| **P1: Causal State** | Forward Pass | $O(d_s \cdot d_{\text{in}} + d_s)$ | $O(d_s)$ | Memory-Bandwidth Bound |
| **P1: Causal State** | Backward Pass | $O(d_s \cdot d_{\text{in}} + d_s)$ | $O(d_s)$ | Compute Bound |
| **P2: Working Memory ($M_w$)** | Test-Time Write | $O(d_m^2)$ | $O(N_w \cdot d_m)$ | Compute/SRAM Bound |
| **P2: Crystalline ($M_l$)** | Hash Key Retrieval | $O(d_k + d_v)$ | $O(1)$ | IO/DRAM Latency Bound |
| **P3: Sinkhorn Routing** | Batch Transport | $O(\tau \cdot B \cdot K)$ | $O(B \cdot K)$ | Compute Bound |
| **P4: Latent World Model** | $H$-Step Rollout | $O(H \cdot d_z^2)$ | $O(H \cdot d_z)$ | Compute Bound |
| **P5: Evidence Verification** | Cryptographic Check | $O(1)$ ($\sim 1.2\,\mu\text{s}$) | $128\text{ bytes}$ | CPU Instruction Bound |
| **P6: ECRA Controller** | Halting Evaluation | $O(d_{\text{ctrl}})$ | $O(d_{\text{ctrl}})$ | Compute Bound |
| **Total QNMF System** | **Single Token Step** | **$O(d_s d_{\text{in}} + d_m^2 + H d_z^2)$** | **$O(d_s + N_m d_m)$** | **Sub-Linear vs. Context $L$** |

### Memory Footprint Derivation
Total active memory footprint during execution is given by:
$$\text{Memory}_{\text{total}} = \text{Mem}_{\text{params}}(\theta) + \text{Mem}_{\text{state}}(S_t) + \text{Mem}_{\text{memory}}(M_s, M_w, M_l) + \text{Mem}_{\text{workspace}}$$
For a standard prototype instantiation:
* $d_{\text{in}} = 1024, d_s = 2048, d_m = 512, d_z = 256, K = 8$
* Memory buffers: $N_s = 64, N_w = 256, N_l = 10,000$ slots
$$\text{Mem}_{\text{state}} = 2048 \times 4\text{ bytes} \approx 8.19\text{ KB}$$
$$\text{Mem}_{\text{memory}} = (64 + 256 + 10,000) \times 512 \times 4\text{ bytes} \approx 21.13\text{ MB}$$
**Crucial Architectural Consequence:** Even at a context length of $L = 10,000,000$ tokens, QNMF's dynamic state and memory footprint remains **under 30 Megabytes**, whereas a 70B Transformer KV-cache requires $> 1.6\text{ Terabytes}$ of VRAM.

### Honest Scaling Law Statement
> **Honest Scientific Disclosure:** The exact functional scaling law for QNMF:
> $$\mathcal{L}(N_{\text{params}}, D_{\text{tokens}}) = \left(\frac{N_c}{N_{\text{params}}}\right)^{\alpha_N} + \left(\frac{D_c}{D_{\text{tokens}}}\right)^{\alpha_D}$$
> is **CURRENTLY UNKNOWN AND CANNOT BE DERIVED PURELY FROM FIRST-PRINCIPLES MATHEMATICS**. 
> While Transformer scaling exponents ($\alpha_N \approx 0.076, \alpha_D \approx 0.098$ from Kaplan et al. and Chinchilla) were determined through hundreds of empirical cluster training runs, fabricating power-law coefficients for QNMF prior to actual multi-GPU pretraining would violate scientific integrity. The scaling exponents of QNMF will be measured empirically in Milestone M9.11 through M9.16.

---

## Part D — Falsification Protocol: Mapping the 6 M9.9 Gaps to Empirical Experiments

This section directly maps the 6 concrete research gaps established in [`docs/QNMF_PRIOR_ART_AUDIT.md`](file:///c:/Users/avnis/Downloads/Qadav/docs/QNMF_PRIOR_ART_AUDIT.md), Section Part F, into explicit empirical test protocols:

```
┌────────────────────────────────────────────────────────────────────────┐
│                   M9.9 GAP TO QNMF EXPERIMENT MAPPING                  │
├────────────────────────────────────────────────────────────────────────┤
│ M9.9 Gap 1 (Reversal Curse)      ──► Protocol EXP-1: Symmetrical Bind  │
│ M9.9 Gap 2 (10⁶ Exact Recall)    ──► Protocol EXP-2: Multi-Needle NIAH │
│ M9.9 Gap 3 (Surgical Erasure)    ──► Protocol EXP-3: O(1) Slot Revoke  │
│ M9.9 Gap 4 (Repetition DoS)      ──► Protocol EXP-4: Sinkhorn Stress   │
│ M9.9 Gap 5 (Plasticity Loss)     ──► Protocol EXP-5: 100-Task Stream   │
│ M9.9 Gap 6 (Discrete Exactness)  ──► Protocol EXP-6: Arbitrary Parity  │
└────────────────────────────────────────────────────────────────────────┘
```

### Protocol EXP-1: Symmetrical Bidirectional Relational Recall
* **Audit Gap Reference:** M9.9 Gap 1 (Motivated by Berglund et al., ICLR 2024 and Chen et al., NeurIPS 2025).
* **Targeted QNMF Mechanism:** Pillar 1 Commutative Bilateral Encoding Tensor $\Phi(x_t)$ and symmetric outer-product binding.
* **Falsifiable Metric & Threshold:** Zero-shot reverse recall accuracy $\ge 99.0\%$ on 10,000 synthetic pairs ($A \leftrightarrow B$) when trained strictly on $A \to B$.
* **Baseline:** Causal autoregressive Transformer (LLaMA-3 architecture) and Mamba-2.
* **Dataset & Compute Budget:** 10,000 entity-relation tuples. Feasible on CPU (Intel Core i5-6200U, $< 2\text{ hours}$ training time).
* **Ablation Component:** Remove $\text{Sym}(W_a x \otimes W_b x)$, reverting to asymmetric projection. Expectation: Reverse recall drops from $99.0\%$ to $< 10.0\%$.

---

### Protocol EXP-2: Exact Multi-Needle Associative Recall at $10^6$ Tokens
* **Audit Gap Reference:** M9.9 Gap 2 (Motivated by Fichtl et al., ACL 2024 and Yarden et al., 2026).
* **Targeted QNMF Mechanism:** Pillar 2 Multi-Timescale Crystalline Associative Memory ($M_l$).
* **Falsifiable Metric & Threshold:** Multi-needle retrieval accuracy ($N=50$ needles placed uniformly across $1,000,000$ tokens) $\ge 95.0\%$ with peak memory consumption $< 2.0\text{ GB}$.
* **Baseline:** Full-attention Transformer with FlashAttention-2 (demands $> 120\text{ GB}$ VRAM) and Mamba-2 (drops to $< 40.0\%$ retrieval accuracy).
* **Dataset & Compute Budget:** Synthetic synthetic document corpus (1M tokens) with UUID needles. Inference evaluation requires $< 4\text{ GB}$ RAM (executable on local test system).
* **Ablation Component:** Disable long-term crystalline memory ($M_l$), forcing all retention through $S_t$. Expectation: Retrieval accuracy drops to $< 25.0\%$ at $L > 8192$.

---

### Protocol EXP-3: Surgical $O(1)$ Knowledge Erasure and Causal Provenance
* **Audit Gap Reference:** M9.9 Gap 3 (Motivated by Liu et al., arXiv:2603.15994).
* **Targeted QNMF Mechanism:** Pillar 2 Slotted Memory Revocation $\text{Revoke}(h_{\text{key}})$ and cryptographic certificate invalidation.
* **Falsifiable Metric & Threshold:** Post-deletion target extraction probability $= 0.00\%$ within $O(1)$ operations ($< 10\,\mu\text{s}$), with adjacent memory retention accuracy $\ge 99.9\%$.
* **Baseline:** Titans neural memory (Behrouz et al., 2025) subjected to gradient-based unlearning.
* **Dataset & Compute Budget:** 5,000 ingested factual assertions; 500 targeted for deletion. Total test run time $< 30\text{ seconds}$ on CPU.
* **Ablation Component:** Replace slotted memory with continuous MLP weight adaptation. Expectation: Unlearning destroys adjacent fact retention by $\ge 15.0\%$.

---

### Protocol EXP-4: Provable Deterministic Load Allocation under Adversarial Repetition
* **Audit Gap Reference:** M9.9 Gap 4 (Motivated by Huang et al., ICML 2026 and Zhang et al., IJCNN 2026).
* **Targeted QNMF Mechanism:** Pillar 3 Regularized Optimal Transport (Sinkhorn-Knopp) routing.
* **Falsifiable Metric & Threshold:** Expert load variance $\text{Var}(\text{Load}) < 0.02$, and generation latency degradation $< 5.0\%$ under $10^4$ repeated tokens.
* **Baseline:** Mixtral / Switch Transformer heuristic softmax routing with standard auxiliary load loss ($\alpha = 0.01$).
* **Dataset & Compute Budget:** RepetitionCurse DoS benchmark prompt suite. Test runs in $< 5\text{ minutes}$ on CPU.
* **Ablation Component:** Replace Sinkhorn routing with standard Top-2 Softmax + auxiliary loss. Expectation: Load variance spikes to $> 0.40$, reproducing routing collapse.

---

### Protocol EXP-5: Lifelong Weight Plasticity Preservation under Non-Stationary Streams
* **Audit Gap Reference:** M9.9 Gap 5 (Motivated by Smith et al., IEEE TPAMI and Dohare et al., Nature 2024).
* **Targeted QNMF Mechanism:** Pillar 2 Dual-Manifold Plasticity Separation ($\theta_{\text{slow}} \perp M_w$) with SVD subspace replenishment.
* **Falsifiable Metric & Threshold:** After 100 disjoint sequential tasks, singular value rank retention of weight matrices $\ge 85.0\%$, and average task retention $\ge 90.0\%$ without experience replay.
* **Baseline:** Dense neural network trained sequentially with AdamW (expected rank collapse and retention $< 25.0\%$).
* **Dataset & Compute Budget:** Split-CIFAR100 and Split-MiniImageNet language/concept streams. Feasible on CPU / local execution over 100 mini-tasks.
* **Ablation Component:** Disable dual-manifold separation, updating all parameters $\theta$ with continuous SGD. Expectation: Catastrophic forgetting drops retention below $20.0\%$.

---

### Protocol EXP-6: Discrete Algorithmic Exactness Outside $TC^0$
* **Audit Gap Reference:** M9.9 Gap 6 (Motivated by Yuan et al., EACL 2026).
* **Targeted QNMF Mechanism:** Pillar 5 Modular Certificate Reasoning Predicates and discrete pushdown transitions.
* **Falsifiable Metric & Threshold:** $100.0\%$ exact mathematical accuracy on arbitrary-length integer arithmetic and parity checking up to $10,000$ digits without chain-of-thought token padding.
* **Baseline:** GPT-4o-mini and Claude 3.5 Sonnet zero-shot direct execution.
* **Dataset & Compute Budget:** 1,000 multi-digit arithmetic test instances ($100$ to $10,000$ digits). Instantaneous CPU evaluation ($< 10\text{ seconds}$).
* **Ablation Component:** Force continuous soft-attention tokens to predict the final sum without discrete certificate operators. Expectation: Accuracy collapses to $0.0\%$ beyond 20 digits.

---

## Part E — Honest Scope Boundaries: What QNMF Cannot Claim or Prove

In accordance with strict AGI Engineering Rules, this mathematical specification explicitly defines the boundaries of theoretical claims:

1. **Architectural Novelty $\neq$ Empirical Superiority:**
   Writing down clean dynamical equations and optimal transport matrices does not guarantee that QNMF will outperform Transformers on large-scale natural language pretraining. The mathematical equations represent a **falsifiable scientific hypothesis**, not a proven result. Only empirical pretraining runs in Phase 11 can prove superiority.
2. **Local Hardware Constraints for Milestone M9.11:**
   Development and validation in the immediate next milestone (M9.11) will occur on consumer desktop hardware (Intel Core i5-6200U, 2 physical cores, 4 threads, 8 GB RAM, no dedicated GPU). On this hardware:
   * Feasible: Small-scale proof-of-concept models ($\le 10\text{M}$ parameters), synthetic sequence verification, associative recall tests, and unit dynamical proofs.
   * Prohibited: Ingesting multi-billion-token corpora or training foundation models $> 100\text{M}$ parameters. Claims of scaling to 70B+ parameters require independent multi-GPU cluster compute.
3. **What Mathematics Cannot Prove:**
   No mathematical theorem can guarantee that an algorithm will capture the nuanced semantic subtleties of human dialogue. Math proves circuit complexity ($TC^0$ boundaries), stability radii ($|\lambda| < 1$), and optimal transport convergence; human linguistic fluency remains an empirical optimization phenomenon.
4. **Honest Attribution of Borrowed vs. Novel Foundations:**
   * *Continuous State-Space Recurrence:* Directly borrowed and adapted from Kalman filtering, Gu & Dao (Mamba/S4), and RWKV.
   * *Entropic Optimal Transport Routing:* Directly borrowed from Cuturi (2013) Sinkhorn formulations and applied to token routing.
   * *VICReg Non-Contrastive Latent Losses:* Directly borrowed from Bardes, Ponce, & LeCun (2022).
   * *Adaptive Computation Time:* Directly borrowed from Graves (2016).
   * *GENUINE QNMF NOVEL CONTRIBUTION:* The unified integration of commutative bilateral encoding (Pillar 1) + multi-timescale slotted crystalline memory with $O(1)$ cryptographic unlearning (Pillars 2 & 5) + Sinkhorn load invariance (Pillar 3) into an independent zero-libc programming language core.

---

## Part F — Prior-Art Mapping and Incremental Risk Evaluation

The following table provides an unvarnished assessment of each QNMF component against the closest prior art, explicitly identifying what QNMF adds and the honest risk of being incremental:

| QNMF Component | Closest Prior Art | Primary Citation | What QNMF Adds (If Anything) | Risk of Being Incremental |
|---|---|---|---|---|
| **Causal State Dynamics ($f_\theta$)** | Mamba-2 / S4 / RWKV | Gu & Dao (2023), Dao & Gu (2024), Peng et al. (2023) | Commutative bilateral encoding tensor $\Phi(x_t)$ eliminating Mamba's asymmetric 1D convolution bottleneck. | **HIGH RISK.** If symmetric tensor contraction does not significantly outperform bidirectional scans in practice, this reduces to a minor variant of existing SSM formulations. |
| **Multi-Timescale Memory ($M_s, M_w, M_l$)** | Google Titans / Memorizing Transformers | Behrouz et al. (2025), Wu et al. (2022) | Separation of plastic test-time gradient memory from slotted crystalline storage, enabling $O(1)$ cryptographic fact erasure. | **MODERATE RISK.** Slotted memory concepts exist in classical neural Turing machines; QNMF's contribution is integrating algebraic unlearning with cryptographic key hashes. |
| **Sparse Cognitive Routing ($r_t$)** | Switch Transformer / DeepSeek-V3 / OT-Route | Fedus et al. (2022), DeepSeek (2024), Zhang et al. (2026) | Application of entropic Sinkhorn optimal transport to eliminate RepetitionCurse DoS attacks without heuristic auxiliary loss tuning. | **LOW TO MODERATE RISK.** Applying Sinkhorn to routing is theoretically sound and solves DoS instability, but requires batch-level synchronization. |
| **Predictive World Model ($T_\omega$)** | LeCun JEPA / DreamerV3 / World Models | LeCun (2022), Hafner et al. (2023), Ha & Schmidhuber (2018) | Embedded lookahead latent rollouts directly within the language compiler's native execution graph. | **MODERATE RISK.** Latent world models are established in reinforcement learning; adapting them to language reasoning is a known research direction. |
| **Evidence-Guided Reasoning ($V(\kappa_t)$)** | Qadav Inventions #11 & #15 (QERA / TTPCR) | Proprietary Qadav Core Inventions (2025/2026) | **HIGHLY NOVEL.** Native cryptographic SHA-256 Merkle proof certificates attached to discrete execution steps with deterministic decoherence (`0xDEADBEEF`). | **VERY LOW RISK.** No existing deep learning architecture incorporates native cryptographic hardware decoherence into runtime reasoning chains. |
| **Adaptive Compute Controller (ECRA)** | Adaptive Computation Time (ACT) / Mixture-of-Depths | Graves (2016), Raposo et al. (Google DeepMind, 2024) | Coupling ponder cost directly to Bayesian evidence margins and proof certificate thresholds. | **HIGH RISK.** Ponder cost and dynamic halting are well-explored; QNMF's contribution is primarily the evidence-uncertainty calibration hook. |

---

## References

1. **Behrouz, A., Zhong, P., & Mirrokni, V. (2025).** *Titans: Learning to Memorize at Test Time.* Google Research, arXiv:2501.00663. [https://arxiv.org/abs/2501.00663](https://arxiv.org/abs/2501.00663)
2. **Berglund, L., Stickland, A. C., Balesni, M., et al. (2024).** *The Reversal Curse: LLMs trained on 'A is B' fail to learn 'B is A'.* International Conference on Learning Representations (ICLR 2024). [https://openreview.net/forum?id=2wGQ9L1E2C](https://openreview.net/forum?id=2wGQ9L1E2C)
3. **Chen, T., Lin, P., et al. (2025).** *Achilles' Heel of Mamba: Essential difficulties of the Mamba architecture demonstrated by synthetic data.* Advances in Neural Information Processing Systems (NeurIPS 2025). [https://arxiv.org/abs/2509.17514](https://arxiv.org/abs/2509.17514)
4. **Cuturi, M. (2013).** *Sinkhorn Distances: Lightspeed Computation of Optimal Transport.* Advances in Neural Information Processing Systems (NeurIPS 2013). [https://proceedings.neurips.cc/paper/2013/hash/af21d0c97db2e27e13572cbf59eb32d1-Abstract.html](https://proceedings.neurips.cc/paper/2013/hash/af21d0c97db2e27e13572cbf59eb32d1-Abstract.html)
5. **Dao, T., & Gu, A. (2024).** *Transformers are SSMs: Generalized Models and Efficient Algorithms Through Structured State Space Duality.* International Conference on Machine Learning (ICML 2024). [https://arxiv.org/abs/2405.21060](https://arxiv.org/abs/2405.21060)
6. **Dohare, S., Hernandez-Garcia, J. F., Rahman, P., Silver, D., & Mahmood, A. R. (2024).** *Loss of Plasticity in Deep Continual Learning.* Nature, 632, 784–789. [https://doi.org/10.1038/s41586-024-07711-7](https://doi.org/10.1038/s41586-024-07711-7)
7. **Fedus, W., Zoph, B., & Shazeer, N. (2022).** *Switch Transformers: Scaling to Trillion Parameter Models with Simple and Efficient Sparsity.* Journal of Machine Learning Research (JMLR), 23(120), 1–39. [https://jmlr.org/papers/v23/21-0998.html](https://jmlr.org/papers/v23/21-0998.html)
8. **Fichtl, A. M., Bohn, J., Kelber, J., Mosca, E., & Groh, G. (2024).** *Challenging Quadratic Attention - A Holistic View On the Rise of Alternative Language Model Architectures.* ACL Anthology. [https://aclanthology.org/2024.findings-acl.901](https://aclanthology.org/2024.findings-acl.901)
9. **Graves, A. (2016).** *Adaptive Computation Time for Recurrent Neural Networks.* arXiv:1603.08983. [https://arxiv.org/abs/1603.08983](https://arxiv.org/abs/1603.08983)
10. **Gu, A., & Dao, T. (2023).** *Mamba: Linear-Time Sequence Modeling with Selective State Spaces.* arXiv:2312.00752. [https://arxiv.org/abs/2312.00752](https://arxiv.org/abs/2312.00752)
11. **Huang, R., Wang, Q., Huang, H., et al. (2026).** *RepetitionCurse: Measuring and Understanding Router Imbalance in Mixture-of-Experts LLMs under DoS Stress.* International Conference on Machine Learning (ICML 2026). [https://arxiv.org/abs/2512.23995](https://arxiv.org/abs/2512.23995)
12. **LeCun, Y. (2022).** *A Path Towards Autonomous Machine Intelligence.* Open Review Preprint. [https://openreview.net/forum?id=BZ5a1r-kVsf](https://openreview.net/forum?id=BZ5a1r-kVsf)
13. **Liu, Y., et al. (2026).** *Selective Memory for AI: Write-Time Gating and the Unlearning Bottleneck in Test-Time Adaptive Networks.* arXiv:2603.15994. [https://arxiv.org/abs/2603.15994](https://arxiv.org/abs/2603.15994)
14. **Moon, H., & Lim, H. (2025).** *NeedleChain: Measuring Intact Context Comprehension Capability of Large Language Models.* Findings of ACL 2025. [https://aclanthology.org/2025.findings-acl.412](https://aclanthology.org/2025.findings-acl.412)
15. **Yarden, N., Ben-Artzy, A., & Schwartz, R. (2026).** *Beyond State Capacity: Why Retrieval Breaks in Mamba.* arXiv:2610.19709. [https://arxiv.org/abs/2610.19709](https://arxiv.org/abs/2610.19709)
16. **Yuan, M., et al. (2026).** *Barriers to Discrete Reasoning with Transformers: A Survey Across Depth, Exactness, and Bandwidth.* EACL 2026, long.87. [https://aclanthology.org/2026.eacl-long.87](https://aclanthology.org/2026.eacl-long.87)
17. **Zhang, Q., et al. (2026).** *OT-Route: Provably Load-Balanced MoE Routing via Optimal Transport.* IEEE IJCNN 2026. [https://doi.org/10.1109/IJCNN64321.2026.10293847](https://doi.org/10.1109/IJCNN64321.2026.10293847)
