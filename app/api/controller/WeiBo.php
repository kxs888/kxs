<?php
namespace app\api\WeiBo;
use think\Controller;
class WeiBo extends Controller {
    
    protected $AppKey = '3349458625';
    protected $App_secret = '648de87da6b7b2496e70b13270e7646f';
    protected $redirect_url = 'www.baidu.com';

    public function __construct(){
        $this->AppKey = $AppKey;
        $this->App_secret = $App_secret;
        $this->redirect_url = $redirect_url;
    }
   /**
    * 获取code
    * @param  $url 
    * @return code state 
    * 同意授权后带着code跳转到 redirect_url
    */
    public function getCode($url){
//         $url = $this->request->post('url');
        $uri = "https://api.weibo.com/oauth2/authorize?client_id=.$AppKey.&redirect_uri=.$redirect_url.&response_type=code&state=STATE";
        $res = json_decode($this->httpGet($uri));
        return $res;
    
    }
    /**
     * 通过code换取access_token
     * @param  $code
     * @return access_token,expires_in,remind_in,uid
     */
    public function getAccessToken($code){
        $url = "https://api.weibo.com/oauth2/access_token?client_id=.$AppKey.&client_secret=.$App_secret.&grant_type=authorization_code&redirect_uri=.$redirect_url.&code=$code";
        $res = json_decode($this->httpGet($url));
        return $res;
    }
    
    public function getTokenInfo($access_token){
        $url = "https://api.weibo.com/oauth2/get_token_info?access_token=$this->access_token";
        $res = json_decode($this->httpGet($url));
        return $res;
    }
    /**
     * 微博授权
     * @param url
     * @return access_token,uid,expir
     */
    public function weiBoAuth($url){
        $code = $this->getCode($url);
        $access_Token = json_encode($this->getAccessToken($code));
        return $access_Token;
        
    }
    
    
    
    
}