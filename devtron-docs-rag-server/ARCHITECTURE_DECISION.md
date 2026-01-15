# Architecture Decision: LLM Processing Location

## 🎯 The Question

**Where should LLM processing happen when integrating with Athena-BE?**

1. **Option A:** RAG API processes LLM (`use_llm=true`)
2. **Option B:** Athena-BE processes LLM (`use_llm=false`) ✅ **RECOMMENDED**

---

## 📊 Detailed Comparison

### Option A: LLM in RAG API (`use_llm=true`)

```
┌──────────┐
│   User   │
└────┬─────┘
     │ "How to deploy apps?"
     ▼
┌─────────────────────────────────┐
│         Athena-BE               │
│  (Has LLM engine)               │
└────┬────────────────────────────┘
     │ POST /search (use_llm=true)
     ▼
┌─────────────────────────────────┐
│      Docs RAG API               │
│  1. Vector search (200ms)       │
│  2. Format context              │
│  3. Call AWS Bedrock ← 💸 LLM #1│
│     (2-3 seconds)               │
│  4. Return enhanced response    │
└────┬────────────────────────────┘
     │ {results: [...], llm_response: "..."}
     ▼
┌─────────────────────────────────┐
│         Athena-BE               │
│  5. Process LLM response        │
│  6. Call LLM again ← 💸💸 LLM #2 │
│     (2-3 seconds)               │
│  7. Return to user              │
└────┬────────────────────────────┘
     │
     ▼
┌──────────┐
│   User   │
└──────────┘

Total Time: ~5-6 seconds
Total Tokens: ~5000 tokens
Total Cost: ~$0.0125 per query
LLM Calls: 2 ❌
```

**Problems:**
- ❌ **Double token consumption** - LLM called twice
- ❌ **Double cost** - Pay for tokens twice
- ❌ **Higher latency** - Two sequential LLM calls
- ❌ **Duplicate logic** - LLM prompting in two places
- ❌ **Less flexible** - Can't easily combine with other sources
- ❌ **Requires AWS credentials** - In RAG API

---

### Option B: LLM in Athena-BE (`use_llm=false`) ✅

```
┌──────────┐
│   User   │
└────┬─────┘
     │ "How to deploy apps?"
     ▼
┌─────────────────────────────────┐
│         Athena-BE               │
│  (Has LLM engine)               │
└────┬────────────────────────────┘
     │ POST /search (use_llm=false)
     ▼
┌─────────────────────────────────┐
│      Docs RAG API               │
│  1. Vector search (200ms)       │
│  2. Return raw results          │
└────┬────────────────────────────┘
     │ {results: [{doc1}, {doc2}, {doc3}]}
     ▼
┌─────────────────────────────────┐
│         Athena-BE               │
│  3. Format context              │
│  4. Combine with other sources  │
│  5. Call LLM once ← 💸 LLM #1   │
│     (2-3 seconds)               │
│  6. Return to user              │
└────┬────────────────────────────┘
     │
     ▼
┌──────────┐
│   User   │
└──────────┘

Total Time: ~3 seconds
Total Tokens: ~3000 tokens
Total Cost: ~$0.0075 per query
LLM Calls: 1 ✅
```

**Benefits:**
- ✅ **Single token consumption** - LLM called once
- ✅ **Half the cost** - Pay for tokens once
- ✅ **Lower latency** - One LLM call
- ✅ **Centralized logic** - All LLM in Athena-BE
- ✅ **More flexible** - Can combine docs with other context
- ✅ **No AWS credentials needed** - In RAG API

---

## 💰 Cost Analysis

### Scenario: 10,000 queries per month

#### Option A (use_llm=true)
```
RAG API LLM calls:    10,000 × 2000 tokens × $0.00125 = $25.00
Athena-BE LLM calls:  10,000 × 3000 tokens × $0.00125 = $37.50
─────────────────────────────────────────────────────────
Total monthly cost:                                $62.50
```

#### Option B (use_llm=false)
```
RAG API LLM calls:    0 × 2000 tokens × $0.00125 = $0.00
Athena-BE LLM calls:  10,000 × 3000 tokens × $0.00125 = $37.50
─────────────────────────────────────────────────────────
Total monthly cost:                                $37.50
```

**Savings: $25/month (40% reduction)** 💰

At scale (100,000 queries/month): **$250/month savings!**

---

## 🚀 Performance Analysis

### Latency Breakdown

#### Option A (use_llm=true)
| Step | Time | Service |
|------|------|---------|
| Vector search | 200ms | RAG API |
| LLM call #1 | 2500ms | RAG API → AWS Bedrock |
| Network transfer | 50ms | RAG API → Athena-BE |
| LLM call #2 | 2500ms | Athena-BE → LLM |
| **Total** | **5250ms** | |

#### Option B (use_llm=false)
| Step | Time | Service |
|------|------|---------|
| Vector search | 200ms | RAG API |
| Network transfer | 50ms | RAG API → Athena-BE |
| LLM call | 2500ms | Athena-BE → LLM |
| **Total** | **2750ms** | |

**Improvement: 2.5 seconds faster (48% reduction)** ⚡

---

## 🔧 Flexibility Comparison

### Option A: Limited Flexibility
```python
# In Athena-BE
response = rag_api.search(query, use_llm=true)
llm_response = response['llm_response']  # Already processed

# Can't easily:
# - Combine with other sources
# - Customize the prompt
# - Add user context
# - Use different LLM models
```

### Option B: Maximum Flexibility ✅
```python
# In Athena-BE
docs = rag_api.search(query, use_llm=false)
other_data = get_other_context()

# Full control:
context = format_context(docs, other_data, user_preferences)
custom_prompt = build_prompt(query, context, user_role)
llm_response = athena_llm.generate(custom_prompt)

# Can:
# ✅ Combine multiple sources
# ✅ Customize prompts per user
# ✅ Add user-specific context
# ✅ Use different LLM models
# ✅ Implement caching strategies
# ✅ Add guardrails and filters
```

---

## 🎯 Decision Matrix

| Criteria | Option A (use_llm=true) | Option B (use_llm=false) |
|----------|------------------------|--------------------------|
| **Token Cost** | ❌ High (2x) | ✅ Low (1x) |
| **Latency** | ❌ Slow (~5s) | ✅ Fast (~3s) |
| **Flexibility** | ❌ Limited | ✅ High |
| **Complexity** | ❌ Duplicate logic | ✅ Centralized |
| **AWS Credentials** | ❌ Required in RAG API | ✅ Not needed |
| **Scalability** | ❌ 2x LLM load | ✅ 1x LLM load |
| **Maintenance** | ❌ Two codebases | ✅ One codebase |
| **Debugging** | ❌ Harder | ✅ Easier |

---

## 📝 Recommendation

### ✅ Use Option B (`use_llm=false`) for Athena-BE Integration

**Reasons:**
1. **40% cost savings** on LLM tokens
2. **48% latency reduction** (2.5s faster)
3. **Better architecture** - Single responsibility principle
4. **More flexible** - Can combine multiple sources
5. **Simpler deployment** - No AWS credentials in RAG API
6. **Easier to maintain** - LLM logic in one place

---

## 🛠️ Implementation Guide

### Step 1: Configure RAG API
```bash
# In devtron-docs-rag-server/.env
# No AWS credentials needed!
POSTGRES_HOST=localhost
POSTGRES_DB=devtron_docs
# ... other DB settings
```

### Step 2: Call from Athena-BE
```python
# In Athena-BE MCP tool
def search_devtron_docs(query: str):
    response = requests.post(
        "http://docs-rag-api:8000/search",
        json={
            "query": query,
            "max_results": 5,
            "use_llm": False  # ← Important!
        }
    )
    return response.json()["results"]

def answer_question(query: str):
    # Get docs
    docs = search_devtron_docs(query)
    
    # Format context
    context = format_docs_for_llm(docs)
    
    # Call LLM once
    prompt = f"Question: {query}\n\nContext:\n{context}\n\nAnswer:"
    answer = athena_llm.generate(prompt)
    
    return answer
```

---

## 🎓 When to Use Option A

Option A (`use_llm=true`) is appropriate when:

1. **Standalone usage** - Not integrating with another LLM service
2. **Simple use case** - Don't need to combine multiple sources
3. **Quick prototyping** - Want immediate LLM responses
4. **Testing** - Validating search quality

**Example use cases:**
- CLI tool for documentation search
- Simple Slack bot without LLM backend
- Internal testing/debugging
- Standalone documentation portal

---

## 📚 Related Documentation

- **MCP Integration Guide**: [MCP_INTEGRATION_GUIDE.md](./MCP_INTEGRATION_GUIDE.md)
- **AWS Credentials**: [AWS_CREDENTIALS_GUIDE.md](./AWS_CREDENTIALS_GUIDE.md)
- **API Examples**: [API_EXAMPLES.md](./API_EXAMPLES.md)
- **Quick Start**: [QUICK_START.md](./QUICK_START.md)

---

## ✅ Final Decision

**For Athena-BE integration: Use `use_llm=false`**

This provides:
- ✅ Lower cost (40% savings)
- ✅ Better performance (48% faster)
- ✅ More flexibility
- ✅ Simpler architecture
- ✅ Easier maintenance

---

**Last Updated:** 2026-01-15

