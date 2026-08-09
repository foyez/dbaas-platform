import http from 'k6/http'
import {check} from 'k6'

const BASE_URL = __ENV.BASE_URL || 'http://foyez.runs.onstackit.cloud'

export const options = {
  stages: [
    {duration: '1m', target: 5},
    {duration: '2m', target: 10},
    {duration: '2m', target: 20},
    {duration: '2m', target: 5},
    {duration: '1m', target: 0},
  ],

  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
}

export default function() {
  const resp = http.get(`${BASE_URL}/api/v1/instances`)

  check(resp, {
    'HTTP status is 200': (r) => r.status === 200,
  })
}
