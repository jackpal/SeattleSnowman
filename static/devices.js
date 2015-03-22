// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

function addIP(ip) {
  modifyActiveUntil(ip, "1h");
}

function subIP(ip) {
  modifyActiveUntil(ip, "-1h");
}

function modifyActiveUntil(ip, delta) {
  post("/modifyActiveUntil", "ip="+ip+"&delta="+delta, refresh);
}

function refresh() {
  location.reload();
}

function post(url, params, callback) {
  var http = new XMLHttpRequest();
  http.open("POST", url, true);

  //Send the proper header information along with the request
  http.setRequestHeader("Content-type", "application/x-www-form-urlencoded");

  http.onreadystatechange = function() {//Call a function when the state changes.
    if(http.readyState == 4 && http.status == 200) {
      callback();
    }
  }
  http.send(params);
}
