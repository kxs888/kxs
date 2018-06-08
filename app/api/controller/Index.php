<?php
namespace app\api\controller;
use app\common\controller\Common;
use think\Db;
use app\api\model\User;

class Index extends Common
{
  
    public function login(){
        $array = array();
        $res = input('post.');
        $user = Db::name('user')->where('phone',$res['phone'])->find();
        if(!$user){
            $array['code'] = -1;
            $array['msg']  = '用户不存在';
        }
         
        if (md5($res['passwd']) != ($user['passwd'])) {
            $array['code'] = -1;
            $array['msg']  = '密码不正确';
        } else {
            $array['code'] = 0;
            $array['msg']  = '登陆成功';
        
        }
             return json_encode($array,JSON_UNESCAPED_UNICODE);
        
        
    }  
        
    public function register(){
        $array = array();
        $user1 = new User();
        $user1->username = input('username');
        $user1->phone = input['phone'];
        $user1->passwd = input('passwd');
        $res = $this->save();
        if(!$res){
            $array['code'] = -1;
            $array['msg'] = '保存失败';
        } else  {
            $array['code'] = 0;
            $array['msg'] = '保存成功';
        }
        return $array;
    }   
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
    
//         $arr =array();
        
//         $member = new Member();
//         $member->phone = $this->request->post('phone');
//         if(!$this->isPhone($member->phone)){
//             $arr['code'] = -1;
//             $arr['msg'] = '注册失败';
//             return $arr;
            
//         }
        
//         $member->username = $this->request->post('username'); 
//         $member->passwd = md5($this->request->post('passwd')) ?? " ";
// //         $member->update_time = date('Y-m-d H:i:s', time());
//         $member->create_time = time();
//         $res = $member->save();
//         if ($res) {
//             $arr['code'] = 0;
//             $arr['msg'] = '注册成功';
//         } else {
//             $arr['code'] = -1;
//             $arr['msg'] = '注册失败';
//         }
        
      
//         return json_encode($arr);
//     }
//     public function isPhone($tel){
//         //正则表达式
     
//         if(strlen($tel) == "11")
//         {
//             //上面部分判断长度是不是11位
//             $n = preg_match_all("/13[123569]{1}\d{8}|15[1235689]\d{8}|188\d{8}/",$tel,$array);
//             /*接下来的正则表达式("/131,132,133,135,136,139开头随后跟着任意的8为数字 '|'(或者的意思)
//              * 151,152,153,156,158.159开头的跟着任意的8为数字
//              * 或者是188开头的再跟着任意的8为数字,匹配其中的任意一组就通过了
//              * /")*/
        
//           if(!empty($array)){
//               return true;
//           }
              
//         }else{
        
//             return false;
//         }
//         /*
//          * 虽然看起来复杂点,清楚理解!
//          * 如果有更好的,可以贴出来,分享快乐!
//          * */
//     }
//    }