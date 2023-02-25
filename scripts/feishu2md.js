// ==UserScript==
// @name         feishu2md
// @namespace    http://github.com/Wsine
// @version      0.1
// @description  download the feishu/lark document in markdown format
// @author       Wsine
// @match        https://*.feishu.cn/*
// @match        https://*.larksuite.com/*
// @icon         https://em-content.zobj.net/thumbs/120/apple/325/notebook_1f4d3.png
// @grant        none
// @run-at       context-menu
// ==/UserScript==

(function() {
  'use strict';

  const redirect_uri = "<dev server>"
  const app_id = "cli_a267ad07c4b85013";
  const doc_url = window.location.href;
  const openapi_auth_link = `https://open.feishu.cn/open-apis/authen/v1/index?redirect_uri=${encodeURIComponent(redirect_uri)}&app_id=${app_id}&state=${encodeURIComponent(doc_url)}`;
  console.log(openapi_auth_link)

  window.open(openapi_auth_link, "_blank");
})();
