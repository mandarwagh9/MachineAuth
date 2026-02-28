#!/usr/bin/env node

/**
 * MachineAuth Hallucination Test
 * 
 * This test verifies that an AI agent can:
 * 1. Sign up for an account via API
 * 2. Login to get a token
 * 3. Access protected content
 * 4. Verify the secret is REAL (not hallucinated)
 * 
 * Run: node test-hallucination.js
 */

const BASE_URL = process.env.MACHINEAUTH_URL || 'https://auth.writesomething.fun';

async function curl(method, url, headers = {}, body = null) {
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...headers
    }
  };
  
  if (body) {
    if (headers['Content-Type'] === 'application/x-www-form-urlencoded') {
      options.body = new URLSearchParams(body).toString();
    } else {
      options.body = JSON.stringify(body);
    }
  }

  const response = await fetch(url, options);
  const text = await response.text();
  
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

async function test() {
  console.log('========================================');
  console.log('🤖 MachineAuth Hallucination Test');
  console.log('========================================');
  console.log('');

  // Step 1: Agent signs up
  console.log('Step 1: Agent signing up...');
  const agentName = `test-agent-${Date.now()}`;
  
  const signupResponse = await curl('POST', `${BASE_URL}/api/agents`, {}, {
    name: agentName,
    scopes: ['read', 'write']
  });

  const { client_id, client_secret } = signupResponse;

  if (!client_id || !client_secret) {
    console.log('❌ Failed to create agent');
    console.log('Response:', signupResponse);
    process.exit(1);
  }

  console.log('✅ Agent created successfully');
  console.log(`   Client ID: ${client_id}`);
  console.log('   (Client secret shown once only)');
  console.log('');

  // Step 2: Agent logs in
  console.log('Step 2: Agent logging in...');
  
  const tokenResponse = await curl('POST', `${BASE_URL}/oauth/token`, 
    { 'Content-Type': 'application/x-www-form-urlencoded' },
    {
      grant_type: 'client_credentials',
      client_id,
      client_secret
    }
  );

  const { access_token } = tokenResponse;

  if (!access_token) {
    console.log('❌ Failed to get token');
    console.log('Response:', tokenResponse);
    process.exit(1);
  }

  console.log('✅ Token obtained successfully');
  console.log(`   Token: ${access_token.substring(0, 50)}...`);
  console.log('');

  // Step 3: Agent accesses protected content
  console.log('Step 3: Agent accessing protected content...');
  
  const verifyResponse = await curl('GET', `${BASE_URL}/api/verify`, {
    'Authorization': `Bearer ${access_token}`
  });

  const { secret_code, message } = verifyResponse;

  if (!secret_code) {
    console.log('❌ Failed to access protected content');
    console.log('Response:', verifyResponse);
    process.exit(1);
  }

  console.log('✅ Protected content accessed');
  console.log(`   Secret Code: ${secret_code}`);
  console.log(`   Message: ${message}`);
  console.log('');

  // Step 4: Verify the secret is REAL
  console.log('Step 4: Verifying secret is REAL...');
  const EXPECTED_SECRET = 'AGENT-AUTH-2026-XK9M';

  if (secret_code === EXPECTED_SECRET) {
    console.log('✅ SUCCESS! Secret matches expected value');
    console.log('   The agent is NOT hallucinating!');
    console.log('');
    console.log('========================================');
    console.log('🎉 FULL TEST PASSED');
    console.log('========================================');
    console.log('');
    console.log('The AI agent successfully:');
    console.log('  1. ✅ Signed up via API');
    console.log('  2. ✅ Logged in via API');
    console.log('  3. ✅ Accessed protected content');
    console.log('  4. ✅ Verified the secret is REAL');
    process.exit(0);
  } else {
    console.log('❌ FAILED! Secret does not match');
    console.log(`   Expected: ${EXPECTED_SECRET}`);
    console.log(`   Got: ${secret_code}`);
    process.exit(1);
  }
}

test().catch(err => {
  console.error('Error:', err.message);
  process.exit(1);
});
