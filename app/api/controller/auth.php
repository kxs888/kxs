<?php
namespace app\api\controller\auth;

use app\admin\common\tool\Weichat;
use app\api\model\User;

class Auth extends Controller
{
    // 验证用户是否微信登陆过
    public function initialize()
    {
        $user_id = session('user_id');
        if (! $user_id) {
            $path = $_SERVER['REQUEST_URI'];
            header('Location:/api/auth/index?path=' . $path);
        }
    }

    public function authCode()
    {
        $url = $this->request->post("url");
        $weichat = new Weichat($appId, $appSecret);
        $code = $weichat->getCode(urlEncode($url));
        return $code;
    }

    public function authUserInfo()
    {
        $url = $this->request->post("url");
        $code = $weichat->getCode(urlEncode($url));
        $weichat = new Weichat($appId, $appSecret);
        $userinfo = $weichat->getUserInfo($code);
        if (empty($userinfo)) {
            $array['code'] = - 1;
            $array['msg'] = '授权失败';
        } else {
            $array['code'] = 0;
            $array['msg'] = '授权成功';
            session_start();
            $user = new User();
            $user['nickname'] = $userinfo['nickname'];
            $user['openid'] = $userinfo['openid'];
            $user['sex'] = $userinfo['sex'];
            $user_id = $user->where('uid', $usrinfo['openid'])->select();
            if (! $user_id) {
                $user->save();
            }
        }
        return $array;
    }
    
    public function getUinfo($url){
        $array = array();
        // 1、准备Scope为snsapi_userinfo的网页授权页面URL；
        $url = urlencode($url);
        $code_path = "https://open.weixin.qq.com/connect/oauth2/authorize?appid={$this->appId}&redirect_uri={$this->url}&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect";
        // 2、用户手动同意授权，获取code；
        
        // 页面将跳转至 redirect_uri/?code=CODE&state=STATE
        if (!isset($GET['code'])) {
            header("Location:{$url}");
        }
        $code = $_GET['code'];
        // 3、通过code换取网页授权access_token。
        $access_token_path = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=$this->appId&secret=$appSecret&code=$code&grant_type=authorization_code";
        $res = $this->requset($access_token_path);
        $access_token = $res['access_token'];
        $openid = $res['openid'];
        $userinfo_path = "https://api.weixin.qq.com/sns/userinfo?access_token=$access_token&openid=$openid&lang=zh_CN";
        $result = $this->request($userinfo_path);
        
    }
}