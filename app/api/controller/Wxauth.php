<?php
namespace app\api\controller;
use app\common\controller\Common;  
use think\Request;
class Wxauth extends Common {
    public $appid = "wxd720fbed60c6455a";
    public $secret = "hzqeV3ZEN1QtO8Y1vjqVrbBd8iEY60";
    public function __construct(){
        header("Access-Control-Allow-Origin:*");
    }
    public function index(){
        return 'index';
    }
    /**
     * 第一步  前端js请求(POST) getWxCode 接口，返回一个url。 继续请求返回的url， 转发到指定url获取code
     * 第二步  请求 getUserInfo 接口  code 为参数，，获取到用户信息
     * @return string
     */
    public function getWxCode() {
        $data = input('post.');
        $redirect = $data['url'];
        $url = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=$this->appid&redirect_uri=$this->redirect_uri&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect";
        return $url;
    }
        
    /**
     * http请求类;
     * @param unknown $url
     * @param string $data
     */
    private function http_request($url,$data = null){
        $ch = curl_init();
        curl_setopt($ch,CURLOPT_URL,$url);
        curl_setopt($ch,CURLOPT_SSL_VERIFYPEER,FALSE);
        curl_setopt($ch,CURLOPT_SSL_VERIFYHOST,FALSE);
        curl_setopt($ch,CURLOPT_RETURNTRANSFER,1);
        if(!empty($data))
        {
            curl_setopt($ch,CURLOPT_POST,1);
            curl_setopt($ch,CURLOPT_POSTFIELDS,$data);
        }
        $outopt = curl_exec($ch);
        curl_close($ch);
        return $outopt;
    }
    /**
     * 获取accessToken
     * @return Ambigous <string, unknown>
     */
    public function get_access_token($code){
        $url = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=$this->appid&secret=$this->$secret&code=$this->code&grant_type=authorization_code";
        $access_token =  $this->http_request($url);
        $accessToken = json_decode($access_token,true);
        $access_token = $accessToken['access_token'];
        $data->openid = $accessToken['openid'];
        $this->access_token = $accessToken['access_token'];
       
        return array('access_token'=>$this->access_token, 'openid'=>$this->openid);
    }

    /**
     * 获取用信息
     * @param unknown $openid
     */
    public function getUserInfo($url){
        $request = Request::instance();
        $url = $request->post('url');
        $code = $this->getWxcode($url);
        $res = $this->get_access_token($code);
        $access_token = $res['access_token'];
        $openid = $res['openid'];
        $url = "https://api.weixin.qq.com/sns/userinfo?access_token=$access_token&openid=$openid&lang=zh_CN";
        $userinfo = $this->http_request($url);
//         if(!empty($userinfo)){
//         $userinfo = json_decode($userinfo,true);
//         $data['openid'] = $userinfo['openid'];
//         $data['nickname'] = $userinfo['nickname'];
//         $data['country'] = $userinfo['country'];
//         $data['sex'] = $userinfo['sex'];
//         $data['city'] = $userinfo['city'];
//         $data['headimgurl'] = $userinfo['headimgurl'];
//         $data['create_time'] = time();
//         $this->save();
//         }
        
        return $userinfo;
    }
    private function httpGet($url) {
        $curl = curl_init();
        curl_setopt($curl, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($curl, CURLOPT_TIMEOUT, 500);
        curl_setopt($curl, CURLOPT_SSL_VERIFYPEER, false);
        curl_setopt($curl, CURLOPT_SSL_VERIFYHOST, false);
        curl_setopt($curl, CURLOPT_URL, $url);
        $res = curl_exec($curl);
        curl_close($curl);
        return $res;
    }
}