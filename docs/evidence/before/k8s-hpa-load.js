import http from 'k6/http';

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 200 },  // 30 秒衝到 200 VU
        { duration: '3m',  target: 200 },   // 維持 3 分鐘,等 HPA 擴容
        { duration: '30s', target: 0 },     // 收尾
      ],
    },
  },
};

export default function () {
  http.get('http://localhost:8080/ping');
}
