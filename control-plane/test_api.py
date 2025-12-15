#!/usr/bin/env python3
"""Simple test script for the Control Plane API."""
import base64
import time
import requests

BASE_URL = "http://localhost:8000"


def test_health():
    """Test health endpoint."""
    print("Testing /health endpoint...")
    response = requests.get(f"{BASE_URL}/health")
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}\n")
    return response.status_code == 200


def test_list_agents():
    """Test listing agents."""
    print("Testing /agents endpoint...")
    response = requests.get(f"{BASE_URL}/agents")
    print(f"Status: {response.status_code}")
    data = response.json()
    print(f"Found {data['total']} agent(s)")
    for agent in data['agents']:
        print(f"  - {agent['hostname']} ({agent['status']})")
    print()
    return response.status_code == 200 and data['total'] > 0


def test_deploy():
    """Test deploying a simple nginx project."""
    print("Testing /deploy endpoint...")
    
    # Create a simple docker-compose.yml
    compose_content = """version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8080:80"
"""
    
    # Base64 encode
    compose_base64 = base64.b64encode(compose_content.encode()).decode()
    
    # Deploy
    response = requests.post(
        f"{BASE_URL}/deploy",
        json={
            "project_name": "test-nginx",
            "compose_file_base64": compose_base64
        }
    )
    
    print(f"Status: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Job ID: {data['job_id']}")
        print(f"Deployment ID: {data['deployment_id']}")
        print(f"Status: {data['status']}")
        print(f"Message: {data['message']}")
        print()
        return True, data['job_id']
    else:
        print(f"Error: {response.text}\n")
        return False, None


def test_status(project_name="test-nginx"):
    """Test getting project status."""
    print(f"Testing /projects/{project_name}/status endpoint...")
    response = requests.get(f"{BASE_URL}/projects/{project_name}/status")
    print(f"Status: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Job ID: {data['job_id']}")
        print(f"Status: {data['status']}")
        print(f"Logs: {data['logs']}")
        print()
        return True
    else:
        print(f"Error: {response.text}\n")
        return False


def test_stop(project_name="test-nginx"):
    """Test stopping a project."""
    print(f"Testing /projects/{project_name}/stop endpoint...")
    response = requests.post(
        f"{BASE_URL}/projects/{project_name}/stop",
        json={}
    )
    print(f"Status: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Job ID: {data['job_id']}")
        print(f"Status: {data['status']}")
        print(f"Message: {data['message']}")
        print()
        return True
    else:
        print(f"Error: {response.text}\n")
        return False


def main():
    """Run all tests."""
    print("=" * 60)
    print("Control Plane API Test Suite")
    print("=" * 60)
    print()
    
    # Test health
    if not test_health():
        print("❌ Health check failed. Is the server running?")
        return
    
    print("✅ Health check passed\n")
    
    # Test list agents
    if not test_list_agents():
        print("❌ No agents connected. Please start an agent first.")
        return
    
    print("✅ Agent listing passed\n")
    
    # Test deploy
    deploy_success, job_id = test_deploy()
    if not deploy_success:
        print("❌ Deployment failed")
        return
    
    print("✅ Deployment queued\n")
    
    # Wait a bit for deployment to process
    print("Waiting 5 seconds for deployment to process...")
    time.sleep(5)
    
    # Test status
    if not test_status():
        print("⚠️  Status check failed")
    else:
        print("✅ Status check passed\n")
    
    # Test stop
    if not test_stop():
        print("❌ Stop failed")
        return
    
    print("✅ Stop command sent\n")
    
    print("=" * 60)
    print("All tests completed!")
    print("=" * 60)


if __name__ == "__main__":
    main()


