import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Assume the base URL of your application
const BASE_URL = 'http://localhost:8080';

// A simple counter for tracking the number of successful requests
const postCounter = new Counter('post_requests');
const getCounter = new Counter('get_requests');

export const options = {
  scenarios: {
    producer: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m',
      exec: 'produceMessages',
    },
    consumer: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m',
      exec: 'consumeMessages',
    },
  },
};

// Function to generate a random string of a given length
function generateRandomString(length) {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
}

// Producer scenario: sends POST requests
export function produceMessages() {
  const messageLength = Math.floor(Math.random() * (300 - 250 + 1)) + 250;
  const message = generateRandomString(messageLength);

  const payload = JSON.stringify({ message: message });
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(`${BASE_URL}/messages`, payload, params);
  if (res.status === 200) {
    postCounter.add(1);
  }
}

// Consumer scenario: sends GET and DELETE requests
export function consumeMessages() {
  // const clientId = uuidv4();
  const clientId = uuidv4();
  // Get a message
  const getRes = http.get(`${BASE_URL}/messages?clientId=${clientId}&count=1`);

  if (getRes.status === 200) {
    const messages = getRes.json();
    if (messages && messages.length > 0) {
      getCounter.add(1);
      const messageId = JSON.parse(getRes.body)[0].messageId

      // Delete the message
      const deleteRes = http.del(`${BASE_URL}/messages?messageId=${messageId}&clientId=${clientId}`);
      check(deleteRes, {
        'consumer: delete status is 200': (r) => r.status === 200,
      });
    }
  }
}

// export function handleSummary(data) {
//   console.log(`Total POST requests: ${data.metrics.post_requests.values.count}`);
//   console.log(`Total GET requests: ${data.metrics.get_requests.values.count}`);
//   return {
//     'stdout': JSON.stringify(data, null, 2),
//     'summary.json': JSON.stringify(data),
//   };
// }
