#!/usr/bin/env python3
"""
Test script for Devtron Documentation API
"""

import requests
import json
import time
import sys

API_URL = "http://localhost:8000"


def print_section(title):
    """Print a section header."""
    print("\n" + "=" * 60)
    print(f"  {title}")
    print("=" * 60)


def test_health():
    """Test health endpoint."""
    print_section("Testing Health Endpoint")
    
    try:
        response = requests.get(f"{API_URL}/health")
        response.raise_for_status()
        
        data = response.json()
        print(f"✅ Status: {data['status']}")
        print(f"✅ Database: {data['database']}")
        print(f"✅ Docs Indexed: {data['docs_indexed']}")
        
        return data['docs_indexed']
        
    except Exception as e:
        print(f"❌ Health check failed: {e}")
        return False


def test_reindex(force=False):
    """Test reindex endpoint."""
    print_section(f"Testing Reindex Endpoint (force={force})")
    
    try:
        response = requests.post(
            f"{API_URL}/reindex",
            json={"force": force},
            timeout=300  # 5 minutes timeout for indexing
        )
        response.raise_for_status()
        
        data = response.json()
        print(f"✅ Status: {data['status']}")
        print(f"✅ Message: {data['message']}")
        print(f"✅ Documents Processed: {data['documents_processed']}")
        print(f"✅ Changed Files: {data['changed_files']}")
        
        return True
        
    except Exception as e:
        print(f"❌ Reindex failed: {e}")
        return False


def test_search(query, use_llm=True, max_results=3):
    """Test search endpoint."""
    print_section(f"Testing Search: '{query}'")
    
    try:
        start_time = time.time()
        
        response = requests.post(
            f"{API_URL}/search",
            json={
                "query": query,
                "max_results": max_results,
                "use_llm": use_llm
            },
            timeout=30
        )
        response.raise_for_status()
        
        elapsed = time.time() - start_time
        data = response.json()
        
        print(f"✅ Query: {data['query']}")
        print(f"✅ Total Results: {data['total_results']}")
        print(f"✅ Response Time: {elapsed:.2f}s")
        
        print("\n📄 Search Results:")
        for i, result in enumerate(data['results'], 1):
            print(f"\n  {i}. {result['title']}")
            print(f"     Source: {result['source']}")
            print(f"     Score: {result['score']:.3f}")
            print(f"     Content: {result['content'][:100]}...")
        
        if use_llm and data.get('llm_response'):
            print("\n🤖 LLM Response:")
            print("-" * 60)
            print(data['llm_response'])
            print("-" * 60)
        
        return True
        
    except Exception as e:
        print(f"❌ Search failed: {e}")
        return False


def main():
    """Run all tests."""
    print("\n🧪 Devtron Documentation API Test Suite")
    print(f"API URL: {API_URL}")
    
    # Test 1: Health check
    docs_indexed = test_health()
    
    # Test 2: Reindex if needed
    if not docs_indexed:
        print("\n⚠️  Documentation not indexed. Running initial indexing...")
        print("⏳ This may take a few minutes...")
        if not test_reindex(force=True):
            print("\n❌ Failed to index documentation. Exiting.")
            sys.exit(1)
    else:
        print("\n✅ Documentation already indexed. Skipping reindex.")
    
    # Test 3: Search queries
    test_queries = [
        "How do I deploy an application?",
        "What is CI/CD pipeline?",
        "How to configure Kubernetes?"
    ]
    
    for query in test_queries:
        # Test with LLM
        test_search(query, use_llm=True, max_results=3)
        time.sleep(1)  # Rate limiting
    
    # Test 4: Search without LLM
    print_section("Testing Search Without LLM")
    test_search("How to deploy?", use_llm=False, max_results=5)
    
    # Summary
    print_section("Test Summary")
    print("✅ All tests completed!")
    print("\nNext steps:")
    print("1. Check the API documentation at http://localhost:8000/docs")
    print("2. Try the interactive API at http://localhost:8000/redoc")
    print("3. Integrate with your MCP tools")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n⚠️  Tests interrupted by user")
        sys.exit(0)
    except Exception as e:
        print(f"\n\n❌ Test suite failed: {e}")
        sys.exit(1)

