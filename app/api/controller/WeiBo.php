<?php
namespace app\api\WeiBo;

class WeiBo {
    
    protected $AppKey = '';
    protected $App_secret = '';
    
    public function __construct(){
        $AppKey = $this->AppKey;
        $App_secret = $this->App_secret;
        
    }
    
    public function getCode($url){
        $url = $this->request->post('url');
        $uri = "https://api.weibo.com/oauth2/authorize?client_id=$AppKey&redirect_uri=$url&state=STATE";
        $res = json_decode($this->httpGet($uri));
        return $res;
    
    }
    
    public function getAccessToken($code){
        $url = "https://api.weibo.com/oauth2/access_token?client_id=$AppKey&client_secret=$App_secret&grant_type=authorization_code&redirect_uri=$this->url&code=$this->code";
        $res = json_decode($this->httpGet($url));
        return $res;
    }
    
    public function getTokenInfo($access_token){
        $url = "https://api.weibo.com/oauth2/get_token_info?access_token=$this->access_token";
        $res = json_decode($this->httpGet($url));
        return $res;
    }
    
    public function weiBoAuth($url){
        $code = $this->getCode($url);
        $access_token = $this->getAccessToken($code);
        $userinfo = json_decode($this->getTokenInfo($access_token));
        return $userinfo;
        
    }
    
    
    
    
}