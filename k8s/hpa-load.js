// k6 壓測腳本,專門用來觸發 HPA。
// 大量併發打 /ping (純 CPU,不碰後端),把 app 的 CPU 拉過 50% requests。
// 跑法: k6 run k8s/hpa-load.js
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
