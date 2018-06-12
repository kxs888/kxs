<?php
namespace app\api\WeiBo;

class WeiBo {
    
    public function getCode($url){
        $url = "https://api.weibo.com/oauth2/authorize?client_id={$AppKey}&response_type=code&redirect_uri={$redirect_uri}";
        $res = json_decode($this->httpGet($url));
        return $res;
    
    }
    
    public function getAccessToken($code){
        $url = "https://api.weibo.com/oauth2/access_token?client_id={$AppKey}&client_secret={$App_secret}&grant_type=authorization_code&redirect_uri=Y{$redirect_uri}&code=$this->code";
        $res = json_decode($this->httpGet($url));
        return $res;
    }
    
    public function getTokenInfo(){
        $res = $this->request->getAccessToken($code);
        $url = "https://api.weibo.com/oauth2/get_token_info";
        $
    }
    
    
    
    
}