<?php
namespace app\api\controller;

use think\Controller;

class Image
{

    public function imageMerge()
    {
        $id = rand(1,3).'jpg';
        
        // 图片一
        $path1 = '../public/static/admin/images/$id';
        // 图片二
        $path2 = '../public/static/admin/images/$id';
        // 创建图片对象 1
        $image1 = imagecreatefromjpeg($path1);
        $image2 = imagecreatefromjpeg($path2);
        // 合成图片
        imagecopymerge($image1, $image2, 160, 650, 0, 0, imagesx($image2), imagesy($image2), 100);
        // 输出合成图片
        // $imagename = rand(1000,9999).'-'.time().'jpg';
        
        $name = '';
        $font = '../public/static/admin/lib/fangzheng.ttf';
        $black = imagecolorallocate($image1, 255, 105, 180); // 字体颜色 RGB
        $fontSize = 20;
        $circleSize = 0;
        $fontBox = imagettfbbox($fontSize, 0, $font, $name);
        imagefttext($image1, $fontSize, $circleSize, 240, 730, $black, $font, $name);
        
        header('Content-Type:image/jpeg');
        imagejpeg($image1, '../public/static/admin/images/a.jpg');
        
        // 释放内存
        imagedestroy($image1);
        imagedestroy($image2);
        
        return "http://kxs.ruohua.club/static/admin/images/a.jpg";
    }
    
    public function imgMerge($str,$date){
        //接收数据
        $str = input('str');
        $str = trim($str);
        $date = input('date');
        //获取底图
        $id = rand(1,20).'png';
        $path = "../public/static/admin/images/$id";
        $image = imagecreatefrompng($path);
        //指定字体样式
        $black1 = imagecolorallocate($image, 240, 66, 52); // 字体颜色 RGB
        $black2 = imagecolorallocate($image, 255, 105, 180); // 字体颜色 RGB
        $font =  '../public/static/admin/lib/fangzheng.ttf';
        $font_size ='20';
        $circle_size = '0';
        $len = strlen($str);
        $fontBox1 = imagettfbbox($fontSize, 0, $font, $str);
        $fontBox2 = imagettfbbox($fontSize, 0, $font, $date);
        //将字体加入图片中
        imagefttext($image, $fontSize, $circleSize, 261, 805, $black1, $font, $str);
        imagefttext($image, 16, $circleSize, 322, 870, $black2, $font, $date);
        $xid = rand(1000,999).time();
        header('Content-Type:image/png');
        
        imagepng($image, "../public/static/admin/img/$xid");
        //释放资源
        imagedestroy($image);
        
        return "http://kxs.ruohua.club/static/admin/img/$xid";
        
        
    }
    
    
    // public function textMerge(){
    // $name = input('post.name');
    // $img = rand(1,5).'.jpg';
    // $path = "..public/static/admin/images/$img";
    // $image = imagecreatefromjpeg($path);
    // dump($image);
    // $black = imagecolorallocate($img, 100, 100,100);//字体颜色 RGB
    // $ttf = '../public/static/admin/lib/fangzheng.ttf';
    // imagefttext($image,16,355,20,20,$black,$ttf,$name);
    // $idimg = rand(1000,9999).'jpg';
    
    // header('Content-Type:image/jpeg');
    // imagejpeg($image,"../public/static/admin/images/$idimg");
    
    // //释放内存
    // imagedestroy($image);
    
    // return "http://kxs.ruohua.club/static/admin/images/$idimg";
    
    // }
}