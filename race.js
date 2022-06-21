import http from "k6/http";

function getRandomNumberBetween(end) {
  return Math.floor(Math.random() * end) + 1;
}

export default function () {
  for (let id = 1; id <= 100; id++) {
    const randInt = getRandomNumberBetween(100);
    let url;
    let tag = {};
    if (randInt % 2 === 0) {
      url = `http://localhost:8000/headers?utm=${id}`;
      tag = {
        name: "utm",
      };
    } else {
      url = `http://localhost:8000/headers?bobby=${id}`;
      tag = {
        name: "bobby",
      };
    }
    http.get(url, {
      tags: tag,
    });
  }
}
