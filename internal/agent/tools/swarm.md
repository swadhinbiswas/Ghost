Spawn a swarm of parallel AI workers to tackle complex tasks simultaneously.
Use this for tasks that benefit from multiple perspectives, parallel research, or comprehensive coverage.

**Strategies:**
- `parallel` (default): Workers tackle different aspects simultaneously (research, planning, edge cases, testing)
- `divide_conquer`: Workers handle sequential phases (research → design → implementation → testing)
- `compare`: Workers produce different approaches to the same problem for comparison

**When to use:**
- Complex multi-faceted problems requiring diverse analysis
- Research tasks where multiple angles are valuable
- Comparing different solution approaches
- Comprehensive code reviews from multiple perspectives
- Generating test cases, documentation, and implementation in parallel

**Parameters:**
- `task`: The main task to decompose and execute
- `workers`: Number of parallel workers (1-5, default: 3)
- `strategy`: Decomposition strategy (parallel|divide_conquer|compare)
- `subtasks`: Optional pre-defined subtasks (if omitted, auto-decomposed)

**Example:**
```json
{"task": "Design a REST API for user management", "workers": 4, "strategy": "parallel"}
```
