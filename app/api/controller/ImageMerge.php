<?php
namespace app\api\controller;

use think\Controller;
use think\Request;

class ImageMerge extends Controller
{
   public  function __construct(){
        
        header('Access-Control-Allow-Origin:*');
        
    }

    public function imgMergeA($str,$date){
        //接收数据
        $request = Request::instance();
        $str = $request->post('str');
        $str = trim($str);
        $date = $request->post('date');
        $time = substr($date,0,4).'年'.substr($date,5,2).'月'.substr($date,8,2).'日';
        //随机获取底图
        $id = rand(1,20).'.png';
        $path = "../public/static/admin/images/imgA/$id";
        $image = imagecreatefrompng($path);
        //指定字体样式
        $black1 = imagecolorallocate($image, 235, 66, 52); // 字体颜色 RGB
        $black2 = imagecolorallocate($image, 107, 37, 18); // 字体颜色 RGB
        $font =  '../public/static/admin/lib/ku.ttf';
        $circle_size = '0';
        $len = strlen($str);
        if($len <= 6){
            $font_size = 36;
            imagefttext($image, $font_size, $circle_size, 255, 805, $black1, $font, $str);
        } elseif($len >= 9) {
            $font_size = 28;
            imagefttext($image, $font_size, $circle_size, 245, 805, $black1, $font, $str);
        }
        //$fontBox1 = imagettfbbox($font_size, 0, $font, $str);
        //$fontBox2 = imagettfbbox($font_size, 0, $font, $date);
        //将字体加入图片中
        //imagefttext($image, $font_size, $circle_size, 255, 805, $black1, $font, $time);
        imagefttext($image, 28, $circle_size, 285, 870, $black2, $font, $time);
        $xid = time().rand(1000,9999).'.png';
        header('Content-Type:image/png');

        imagepng($image, "../public/static/admin/img/$xid");
        //释放资源
        imagedestroy($image);

        return "http://kxs.ruohua.club/static/admin/img/$xid";


    }

    public function imgMergeB(){
        $request = Request::instance();
        $str = $request->post('str');
        $str = trim($str);
        $date = Request::instance()->post('date');
        $time = substr($date,0,4).'年'.substr($date,5,2).'月'.substr($date,8,2).'日';
        //随机获取底图
        $id = rand(1,3).'.jpg';
        $path = "../public/static/admin/images/imgB/$id";
        $image = imagecreatefromjpeg($path);
        //指定字体样式
        $black = imagecolorallocate($image, 3, 25, 154); // 字体颜色 RGB
        $font =  '../public/static/admin/lib/textB.TTF';
        //         $font_size = 33;
        $circle_size = '0';
        $len = strlen($str);
        if($len <= 6){
            imagefttext($image, 26, $circle_size, 374, 281, $black, $font, $str);
        } elseif($len >= 9 ) {
            imagefttext($image, 26, $circle_size, 350, 281, $black, $font, $str);
        }
        //         $fontBox1 = imagettfbbox($font_size, 0, $font, $str);
        //         $fontBox2 = imagettfbbox($font_size, 0, $font, $date);
        //将字体加入图片中
        //         imagefttext($image, $font_size, $circle_size, 374, 281, $black, $font, $str);
        imagefttext($image, 20, $circle_size, 286, 723, $black, $font, $time);
        $xid = time().rand(1000,9999).'.jpg';
        header('Content-Type:image/jpeg');

        imagejpeg($image, "../public/static/admin/img/$xid");
        //释放资源
        imagedestroy($image);

        return "http://kxs.ruohua.club/static/admin/img/$xid";
    }
        
    public function yiliImg(){
        $request = Request::instance();
        $str = $request->post('str');
        $str = trim($str);
         
        //随机获取底图
        //         $id = rand(1,3).'.jpg';
        $id = '1.jpg';
        $path = "../public/static/admin/images/imgYili/$id";
        $image = imagecreatefromjpeg($path);
        //指定字体样式
        $black = imagecolorallocate($image, 188, 26, 30); // 字体颜色 RGB
        $font =  '../public/static/admin/lib/yili.ttf';
        //         $font_size = 33;
        $circle_size = '0';  
        $len = strlen($str);
        if($len <= 6){
            imagefttext($image, 30, $circle_size, 268, 720, $black, $font, $str);
        } elseif($len >= 9 ) {
            imagefttext($image, 26, $circle_size, 350, 281, $black, $font, $str);
        }
         
        //将字体加入图片中
         
         
        $xid = time().rand(100,999).'.jpg';
        header('Content-Type:image/jpeg');
    
        imagejpeg($image, "../public/static/admin/img/$xid");
        //释放资源
        imagedestroy($image);
    
        return 'https://html5.bjsidao.com/static/admin/img/'."$xid";
    
    }


}