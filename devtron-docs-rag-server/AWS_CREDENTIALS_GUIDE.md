# AWS Credentials Configuration Guide

## 🎯 Do You Need AWS Credentials?

### ❌ You DON'T need AWS credentials if:
- You're using `use_llm=false` in search requests (recommended for Athena-BE)
- You only want vector search results
- Your calling application (like Athena-BE) handles LLM processing

### ✅ You DO need AWS credentials if:
- You're using `use_llm=true` in search requests
- You want the RAG API to generate LLM-enhanced responses
- You're using this API standalone without another LLM service

---

## 🔐 AWS Bedrock Authentication Methods

The RAG API uses AWS Bedrock for LLM functionality. Boto3 (AWS SDK) supports multiple authentication methods:

### Method 1: Environment Variables (Docker/Production)

**Best for:** Docker containers, CI/CD, production deployments

```bash
# In .env file or docker-compose.yml
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

**Docker Compose Example:**
```yaml
services:
  docs-rag-api:
    image: devtron-docs-rag-server:latest
    environment:
      - AWS_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
```

**Pros:**
- ✅ Explicit and clear
- ✅ Works in any environment
- ✅ Easy to configure in Docker

**Cons:**
- ❌ Credentials in environment (use secrets management in production)
- ❌ Need to rotate keys manually

---

### Method 2: AWS Profile (Local Development)

**Best for:** Local development, testing

```bash
# In .env file
AWS_REGION=us-east-1
AWS_PROFILE=default
```

This uses credentials from `~/.aws/credentials`:
```ini
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

**Pros:**
- ✅ No credentials in code/env files
- ✅ Easy to switch between profiles
- ✅ Standard AWS CLI workflow

**Cons:**
- ❌ Requires AWS CLI configured
- ❌ Doesn't work well in Docker

---

### Method 3: IAM Role (Production on AWS)

**Best for:** Production deployments on AWS (ECS, EKS, EC2)

**No configuration needed in .env!** Just attach an IAM role to your service.

**IAM Policy Example:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-haiku-20240307-v1:0",
        "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
      ]
    }
  ]
}
```

**For ECS:**
```json
{
  "taskRoleArn": "arn:aws:iam::123456789012:role/DevtronDocsRAGRole"
}
```

**For EKS:**
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: devtron-docs-rag
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/DevtronDocsRAGRole
```

**Pros:**
- ✅ Most secure (no credentials in code)
- ✅ Automatic credential rotation
- ✅ Fine-grained permissions
- ✅ AWS best practice

**Cons:**
- ❌ Only works on AWS infrastructure
- ❌ Requires IAM setup

---

## 🔧 How the API Uses Credentials

The API initializes AWS Bedrock client in `api.py`:

```python
# From api.py (lines 75-85)
try:
    bedrock_runtime = boto3.client(
        service_name='bedrock-runtime',
        region_name=aws_region,  # From AWS_REGION env var
        config=Config(read_timeout=300)
    )
    logger.info("AWS Bedrock initialized for LLM responses")
except Exception as e:
    logger.warning(f"AWS Bedrock not available: {e}. LLM responses will be disabled.")
    bedrock_runtime = None
```

**Boto3 credential resolution order:**
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. AWS profile (`AWS_PROFILE` or `~/.aws/credentials`)
3. IAM role (if running on AWS)
4. Instance metadata (EC2)

If none are found, `bedrock_runtime` will be `None` and LLM features will be disabled.

---

## 🧪 Testing AWS Credentials

### Test 1: Check if credentials are configured
```bash
# Using AWS CLI
aws sts get-caller-identity

# Expected output:
{
    "UserId": "AIDAI...",
    "Account": "123456789012",
    "Arn": "arn:aws:iam::123456789012:user/your-user"
}
```

### Test 2: Test Bedrock access
```bash
# List available models
aws bedrock list-foundation-models --region us-east-1

# Test invoke (requires permissions)
aws bedrock-runtime invoke-model \
  --model-id anthropic.claude-3-haiku-20240307-v1:0 \
  --body '{"anthropic_version":"bedrock-2023-05-31","max_tokens":100,"messages":[{"role":"user","content":"Hello"}]}' \
  --region us-east-1 \
  output.json
```

### Test 3: Test RAG API with LLM
```bash
# Start the API
docker-compose up -d

# Search with LLM
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "test",
    "use_llm": true
  }'

# If credentials work: You'll get llm_response
# If credentials fail: llm_response will contain error message
```

---

## 🚨 Troubleshooting

### Error: "AWS Bedrock not available"
**Cause:** No AWS credentials configured or invalid credentials

**Solution:**
1. Check environment variables: `echo $AWS_ACCESS_KEY_ID`
2. Check AWS profile: `aws configure list`
3. Test credentials: `aws sts get-caller-identity`

### Error: "AccessDeniedException"
**Cause:** Credentials valid but missing Bedrock permissions

**Solution:**
Add `bedrock:InvokeModel` permission to your IAM user/role:
```json
{
  "Effect": "Allow",
  "Action": "bedrock:InvokeModel",
  "Resource": "arn:aws:bedrock:*::foundation-model/*"
}
```

### Error: "ModelNotFoundError"
**Cause:** Model not available in your region or account

**Solution:**
1. Check available models: `aws bedrock list-foundation-models --region us-east-1`
2. Request model access in AWS Console → Bedrock → Model access
3. Use a different model ID

---

## 📋 Quick Setup Checklist

### For Athena-BE Integration (Recommended)
- [ ] No AWS credentials needed
- [ ] Use `use_llm=false` in all requests
- [ ] Let Athena-BE handle LLM processing

### For Standalone API with LLM
- [ ] Choose authentication method (env vars, profile, or IAM role)
- [ ] Configure AWS credentials
- [ ] Set `AWS_REGION` environment variable
- [ ] Test credentials with `aws sts get-caller-identity`
- [ ] Request Bedrock model access in AWS Console
- [ ] Test with `use_llm=true` search request

---

## 🔒 Security Best Practices

1. **Never commit credentials** to version control
2. **Use IAM roles** in production (not access keys)
3. **Rotate access keys** regularly if using them
4. **Use least privilege** - only grant `bedrock:InvokeModel` permission
5. **Use AWS Secrets Manager** for storing credentials in production
6. **Enable CloudTrail** to audit Bedrock API calls
7. **Set up billing alerts** to monitor LLM usage costs

---

## 💰 Cost Considerations

AWS Bedrock charges per token:

| Model | Input (per 1K tokens) | Output (per 1K tokens) |
|-------|----------------------|------------------------|
| Claude 3 Haiku | $0.00025 | $0.00125 |
| Claude 3 Sonnet | $0.003 | $0.015 |

**Example:** 1000 searches with LLM (avg 3000 tokens each):
- Haiku: ~$3.75
- Sonnet: ~$45

**Recommendation:** Use `use_llm=false` and process in Athena-BE to avoid double costs!

---

**Last Updated:** 2026-01-15

