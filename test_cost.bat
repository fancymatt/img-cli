@echo off
echo Testing cost confirmation feature
echo.
echo Test 1: Small workflow (should not require confirmation)
echo Command: img-cli.exe workflow outfit-swap ./outfits/test.jpg --test "subject1" --variations 2
echo Expected: 1 subject x 1 outfit x 1 style x 2 variations = 2 images = $0.08
echo.

echo Test 2: Medium workflow (should show cost but not require confirmation)
echo Command: img-cli.exe workflow outfit-swap ./outfits/test.jpg --test "subject1 subject2 subject3" --variations 5
echo Expected: 3 subjects x 1 outfit x 1 style x 5 variations = 15 images = $0.60
echo.

echo Test 3: Large workflow (should require confirmation)
echo Command: img-cli.exe workflow outfit-swap ./outfits/batch/ --test "subject1 subject2" --variations 5
echo If batch has 10 outfits: 2 subjects x 10 outfits x 1 style x 5 variations = 100 images = $4.00
echo.

echo Test 4: Very large workflow (should require confirmation)
echo Command: img-cli.exe workflow outfit-swap ./outfits/batch/ --test "s1 s2 s3 s4 s5" --variations 3
echo If batch has 10 outfits: 5 subjects x 10 outfits x 1 style x 3 variations = 150 images = $6.00
echo.

echo Test 5: Skip confirmation with --no-confirm flag
echo Command: img-cli.exe workflow outfit-swap ./outfits/batch/ --test "s1 s2 s3" --variations 5 --no-confirm
echo Should proceed without asking for confirmation
echo.

echo Test 6: Custom cost via environment variable
echo set IMG_CLI_CONFIRM_THRESHOLD=1.00
echo Command: img-cli.exe workflow outfit-swap ./outfits/test.jpg --test "subject1" --variations 30
echo Should require confirmation at lower threshold (30 images = $1.20)