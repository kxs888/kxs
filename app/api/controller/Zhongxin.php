<?php
namespace app\api\controller;

use think\Controller;
use think\Request;

class Zhongxin extends Controller
{

    public function __construct()
    {
        header("Access-Control-Allow-Origin:*");
    }

    public function index()
    {
        var_dump(gd_info());
        // return 'index';
    }
    
    // touxiang头像图片地址
    // bigImgPath 底图大图图片地址
    public function showImg($touxiang, $bigImgPath)
    {
        $photo = $this->resize_img($touxiang); // 将头像裁剪成原型
        $imgurl = "";
        if (! empty($bigImgPath) && ! empty($photo)) {
            $imgurl = $this->get_page($bigImgPath, $touxiang);
        }
        return $imgurl;
    }
    
    // 图像裁剪
    public function resize_img($url, $path = './')
    {
        $imgname = $path . uniqid() . '.jpg';
        $file = $url;
        list ($width, $height) = getimagesize($file); // 获取原图尺寸
        $percent = (110 / $width);
        // 缩放尺寸
        $newwidth = $width * $percent;
        $newheight = $height * $percent;
        $src_im = imagecreatefromjpeg($file);
        $dst_im = imagecreatetruecolor($newwidth, $newheight);
        imagecopyresized($dst_im, $src_im, 0, 0, 0, 0, $newwidth, $newheight, $width, $height);
        imagejpeg($dst_im, $imgname); // 输出压缩后的图片
        imagedestroy($dst_im);
        imagedestroy($src_im);
        return $imgname;
    }

    public function get_page($bigImgPath, $touxiang)
    {
        $font = "/data/wwwroot/tengshi.bjsidao.com/20180104lulu/ditu/fangzheng.ttf"; // 字体
        $img = imagecreatefromstring(file_get_contents($bigImgPath));
        $width = imagesx($img);
        $height = imagesy($img);
        
        // 文字合成
        $black = imagecolorallocate($img, 91, 51, 14); // 字体颜色 RGB
        $fontSize = 20; // 字体大小
        $circleSize = 0; // 旋转角度
        $nicheng = '';
        $name = "-" . $nicheng . "的绝配新主义-";
        $fontBox = imagettfbbox($fontSize, 0, $font, $name);
        imagefttext($img, $fontSize, $circleSize, ceil(($width - $fontBox[2]) / 2), 792, $black, $font, $name);
        
        // //人物照片合成
        // $path_1 = $photo;
        // $qCodeImg1 = imagecreatefromstring(file_get_contents($path_1));
        // list($qCodeWidth1, $qCodeHight1, $qCodeType1) = getimagesize($path_1);
        // imagecopymerge($img, $qCodeImg1, 48, 90, 0, 0, $qCodeWidth1, $qCodeHight1, 100);
        
        // 头像和底图合成
        $qCodeImg = imagecreatefromstring(file_get_contents($touxiang));
        list ($qCodeWidth, $qCodeHight, $qCodeType) = getimagesize($touxiang);
        imagecopymerge($img, $qCodeImg, 64, 127, 0, 0, $qCodeWidth, $qCodeHight, 100);
        
        list ($bgWidth, $bgHight, $bgType) = getimagesize($bigImgPath);
        $imgname = rand(10000, 99999) . '_' . time() . '.jpg';
        header('Content-Type:image/jpg');
        imagejpeg($img, "/data/wwwroot/tengshi.bjsidao.com/20180104lulu/img/" . $imgname); // 输出图像
        
        imagedestroy($qCodeImg);
        imagedestroy($img); // 释放内存
        return "http://tengshi.bjsidao.com/20180104lulu/img/" . $imgname;
    }
}
