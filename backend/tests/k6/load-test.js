import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Rate } from 'k6/metrics';

// Define custom metrics
const errorRate = new Rate('errors');

// Configuration for different test scenarios
export const options = {
  // Base configuration for development testing
  dev: {
    vus: 10,  // 10 virtual users
    duration: '30s',
    thresholds: {
      http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
      errors: ['rate<0.1'],            // Error rate should be below 10%
    },
  },
  
  // Load testing configuration
  load: {
    stages: [
      { duration: '1m', target: 50 },   // Ramp up to 50 users over 1 minute
      { duration: '3m', target: 50 },   // Stay at 50 users for 3 minutes
      { duration: '1m', target: 0 },    // Ramp down to 0 users over 1 minute
    ],
    thresholds: {
      http_req_duration: ['p(95)<1000'],  // 95% of requests should be below 1s
      errors: ['rate<0.05'],              // Error rate should be below 5%
    },
  },
  
  // Stress testing configuration
  stress: {
    stages: [
      { duration: '1m', target: 100 },    // Ramp up to 100 users over 1 minute
      { duration: '3m', target: 100 },    // Stay at 100 users for 3 minutes
      { duration: '2m', target: 200 },    // Ramp up to 200 users over 2 minutes
      { duration: '3m', target: 200 },    // Stay at 200 users for 3 minutes
      { duration: '1m', target: 0 },      // Ramp down to 0 users
    ],
    thresholds: {
      http_req_duration: ['p(95)<2000'],  // 95% of requests should be below 2s
      errors: ['rate<0.1'],               // Error rate should be below 10%
    },
  },
  
  // Spike testing configuration
  spike: {
    stages: [
      { duration: '30s', target: 10 },    // Warm up with 10 users
      { duration: '1m', target: 300 },    // Spike to 300 users
      { duration: '1m', target: 10 },     // Scale back down
      { duration: '30s', target: 0 },     // Ramp down to 0
    ],
    thresholds: {
      http_req_duration: ['p(95)<5000'],  // 95% of requests should be below 5s during spike
      errors: ['rate<0.15'],              // Error rate should be below 15%
    },
  },
  
  // Default to development testing
  scenarios: {
    default: {
      executor: 'ramping-vus',
      gracefulRampDown: '30s',
      stages: [
        { duration: '30s', target: 20 },  // Ramp up to 20 users over 30s
        { duration: '1m', target: 20 },   // Stay at 20 users for 1 minute
        { duration: '30s', target: 0 },   // Ramp down to 0 users
      ]
    }
  }
};

// Pre-generated test data
const userCredentials = new SharedArray('users', function () {
  return [
    { email: 'test1@example.com', password: 'Password123!' },
    { email: 'test2@example.com', password: 'Password123!' },
    { email: 'test3@example.com', password: 'Password123!' },
    // Add more test users as needed
  ];
});

// Global variables
const BASE_URL = __ENV.API_URL || 'http://localhost:8000';
let authToken = '';

// Helper function to get a random user from the array
function getRandomUser() {
  return userCredentials[Math.floor(Math.random() * userCredentials.length)];
}

// Scenario setup
export function setup() {
  // Perform login to get auth token for the test
  const user = getRandomUser();
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: user.email,
    password: user.password
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const tokenData = loginRes.json();
  return { token: tokenData.token };
}

// Default function that is executed for each virtual user
export default function(data) {
  // Get auth token from setup
  const authToken = data.token;
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`
  };
  
  // Group: Public endpoints
  {
    // Health check
    let healthRes = http.get(`${BASE_URL}/health`);
    check(healthRes, {
      'health check status is 200': (r) => r.status === 200,
    }) || errorRate.add(1);

    sleep(1);
  }
  
  // Group: Authenticated endpoints
  {
    // Get current user profile
    let profileRes = http.get(`${BASE_URL}/auth/me`, { headers: headers });
    check(profileRes, {
      'get profile status is 200': (r) => r.status === 200,
      'profile has correct data': (r) => r.json().hasOwnProperty('email'),
    }) || errorRate.add(1);
    
    sleep(1);
    
    // Get list of courses
    let coursesRes = http.get(`${BASE_URL}/api/courses`, { headers: headers });
    check(coursesRes, {
      'get courses status is 200': (r) => r.status === 200,
      'courses response is array': (r) => Array.isArray(r.json()),
    }) || errorRate.add(1);
    
    sleep(2);
  }
}

// Teardown logic (optional)
export function teardown(data) {
  // Cleanup if needed
  console.log('Load test completed.');
} 