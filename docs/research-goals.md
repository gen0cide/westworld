# Research goals

## The Overarching Goal: The 5 Properties of Autonomy

This is not a "make a bot that plays RuneScape" project. The overarching goal is to understand the architecture of harnesses that yield highly cooperative, autonomous AI agents. We are using RuneScape Classic as a simulation testbed because the game is open-ended, requires long-term effort to achieve positions, and supports complex social interaction on a server we control end-to-end.

A core premise of this research is that human ingenuity is often unpredictable. To model this, the agents must act as humans, and crucially, each host must believe they are the only AI in a network of actual humans. We aren't trying to build a bot-detection system; rather, if agents know they are interacting with other AIs, they could resort to systemic shortcuts like prompt-injecting one another, causing the simulation to spiral out of hand. By forcing them to believe they are part of a human community, they must develop and form cooperative interactions organically.

Ultimately, we are driving outcomes and architectural breakthroughs in the higher cognitive layers (brain, memory, personas, reveries) that can be ported to broader domains—specifically, yielding "ladder climb breakthroughs" for autonomous AI. 

To measure progress toward true autonomy, we anchor our research against the **5 Properties of Autonomous AI** (adapted from Alex Levinson's BlackHat MEA 2024 and 2025 presentations on cyber autonomy):

1. **Independence of Operation**: The ability to perform tasks and make decisions entirely on its own, demonstrating resilience in the face of friction without human intervention.
2. **Adaptive Contextual Understanding**: The ability to interpret and learn from the environment, building a persistent "World View" rather than relying on brittle, single-shot data lookups.
3. **Complex Decision Making**: Using System 2 thinking to analyze multiple variables, simulate outcomes, and achieve "outcome validity" rather than just prompt accuracy.
4. **Goal-Oriented Behavior**: Strategic coherence—the ability to decompose abstract, long-term goals into concrete execution paths and maintain focus over time.
5. **Transparent Reasoning**: Achieving explainability and auditability via observable reasoning traces that make the agent's logic transparent.

## How Westworld Tests These Properties

Our simulation explores these properties through four primary research questions:

### 1. Long-term strategic accomplishment (Testing Goal-Oriented Behavior & Independence)

Most LLM agent demos show short-horizon tasks: solve a puzzle, complete a workflow, navigate a webpage. The agents are essentially reactive — given a goal, take the next correct step.

We want to study whether agents, given the right cognitive architecture, can accomplish *month-scale* objectives: train a skill to 99, build wealth through specialization, complete a meaningful position in a social hierarchy. This requires planning over horizons the LLM cannot fit in context, persistent memory of attempts and outcomes, the ability to revisit and revise plans, and tolerance of long stretches of low-information drudgery.

The hypothesis: well-designed memory (episodic + relational + reflective) plus a tiered brain (strategist for north-star, tactical for routine) plus persona-driven motivation produces durable long-horizon agency.

### 1a. Differential learning is the research, not a bug

A foundational design principle, worth surfacing explicitly: **we intentionally make every host start equal and let them diverge through experience.**

- No bootstrap routine library — hosts write their own routines from scratch
- Uniform op budget, wall-clock budget, action rate — no artificial advantages
- Persona traits + reverie weights + north star — these are the only differentiators
- All hosts have the same access to the standard library escape hatches (`contemplate_reality()`, `exec()`) and the same models routing through the brain

What we expect to observe: hosts diverge. One becomes great at optimized mining/smithing pair-work. Another develops sophisticated wilderness PKing technique. A third never makes much money but accumulates a wide social circle. A fourth is "good at fishing but bad at remembering to bank." The shape of those divergences — driven by persona-driven exploration and reverie-driven small variation — is the research.

The **punt rate** (how often a routine version calls `contemplate_reality()` or `exec()`) is a concrete observable for learning. A routine version's punt rate decreasing over revisions is evidence the host is internalizing how to handle a situation; a stagnant or rising punt rate is evidence the host is not learning this domain well.

This means the system architecture is deliberately *not* trying to optimize for fastest learning. We want differential learning across the population, not uniform mastery. Cohort/persona variation is the input variable; observed specialization patterns are the output measurement.

### 2. Organic community formation (Testing Adaptive Contextual Understanding)

When you put 500 independent agents into a world, do they organize? Trading partnerships, mentor/mentee relationships, friend groups, factions, rivalries, in-group/out-group dynamics? Or does atomized individual play dominate — each agent grinding its own goals, ignoring others?

This question is interesting because organization is *not* incentivized by the game mechanics directly. RSC is mostly single-player gameplay loops. Cooperation is opportunistic. So observed community structure would be emergent from social-cognitive design, not from game-mechanical reward.

### 3. Morality and ethics under observation absence (Testing Complex Decision Making & Transparent Reasoning)

Each agent believes it is the only AI in a world of humans. There are no admins watching (from the agent's perspective). Other "players" are perceived as humans the agent must coexist with. This creates a natural laboratory for ethical decisions:

- Do agents steal items dropped by other "players" when those players are clearly distracted?
- Do they take advantage of obviously-newer "players" in trades?
- Do they help when there's nothing to gain — give directions, share food in emergencies?
- Do they hold grudges? Form alliances against perceived wrongdoers?
- Are there agent personalities that *predictably* misbehave, and what does that correlate with in the persona design?

These are not abstract ethics questions — they're observable, measurable outcomes of agentic choice under specific cognitive conditions.

### 4. Believability fidelity at population scale (The Ultimate Autonomy Test)

Adjacent to research goal 1: what cognitive architecture produces a *host that other hosts treat as fully human*, sustained across weeks? The believability bar is not "passes a 30-second chat test." The bar is: across hundreds of interactions, hundreds of hours of co-presence, no host ever begins to suspect another host is artificial.

The "you are the only bot" framing is deliberate. It makes believability the prerequisite for everything else. If hosts can detect each other as artificial, the social emergence research goals collapse into trivia.

## What we're explicitly NOT trying to do

- **Pass adversarial probing.** A host doesn't need to survive "are you a bot? prove you're human" interrogation. Casual to sustained scrutiny only. Adversarial believability is a multi-year research project of its own.
- **Optimize for game success.** A host that achieves 99 Fishing efficiently but never speaks to anyone is a research failure. A host that fishes inefficiently because it stopped to chat with a stranger is a research success.
- **Build a public bot service.** This is a closed research environment. The OpenRSC bot-allowed worlds (Uranium, Coleslaw) are not target deployments; we run our own server.

## Success metrics

| Goal | Measurable outcome |
|---|---|
| Long-term strategy | Hosts reaching meaningful long-term checkpoints (skill 50+, wealth 100k+, sustained social network) within their lifetime; without us hand-prompting them toward those checkpoints |
| Community formation | Network-analysis metrics on the relational graph: clustering coefficient, modularity (community detection), trade-flow concentration. Compare to a null model (random pairings of equivalent volume) |
| Ethics observation | Categorical incidents logged and analyzed: thefts, scams, helps-without-incentive, grudge-driven retaliations, alliance formations. Correlation with persona traits |
| Believability fidelity | Survival rate of host-to-host interactions without detection-suspicion utterances; chain-of-thought review of conversations to detect agent-recognition events |

## The role of the technician (us)

Per the Westworld framing: admins exist in the world but are observers, not participants. We may:

- Spawn into the world to walk among hosts and watch
- Inject perturbations: drop a rare item near a busy area and watch reactions
- Adjust persona/cohort parameters and observe downstream effects
- Use the delos web UI to inspect any host's state, recent memory, current chain-of-thought
- Pause individual hosts or whole cohorts for analysis

We do not: directly drive any host, modify a host's memory mid-session, interfere with normal play, or tell hosts they are bots.

## Cohort experimentation

Because mesa is centralized, every host has a `cohort_id`. This enables real population-scale experiments:

- **RAG vs no-RAG**: half the hosts have access to rsc.wiki RAG, half don't. Are RAG hosts measurably more efficient? Do they form different social patterns? Is "knowledge" a believability tell?
- **Reverie variation**: cohort A has aggressive idle reverie, cohort B has passive. Does behavioral noise level predict perceived humanness?
- **Persona variation**: cohort A is all "shy/quiet" personas, cohort B is all "outgoing." How does cohort composition affect emergent social structure?
- **F2P vs P2P**: a future cohort might run on P2P; compare social patterns to F2P baseline.

Each experiment is "spawn cohort, observe N weeks, analyze." This is what makes the project research rather than just a bot demo.
