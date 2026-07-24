/**
 * Converts an image file to WebP format using HTML5 Canvas.
 * If the file is not an image, it returns the original file untouched.
 * 
 * @param file The file to convert
 * @param maxResolution The maximum width or height of the output image
 * @returns A Promise that resolves to the new WebP File (or the original file if not an image)
 */
export const convertToWebP = (file: File, maxResolution = 1920): Promise<File> => {
  return new Promise((resolve, reject) => {
    // If not an image, return original file
    if (!file.type.startsWith("image/")) {
      return resolve(file);
    }

    const reader = new FileReader();
    reader.readAsDataURL(file);
    
    reader.onload = (event) => {
      const img = new Image();
      img.src = event.target?.result as string;
      
      img.onload = () => {
        let width = img.width;
        let height = img.height;

        // Resize proportionally if it exceeds maxResolution
        if (width > maxResolution || height > maxResolution) {
          if (width > height) {
            height = Math.round((height * maxResolution) / width);
            width = maxResolution;
          } else {
            width = Math.round((width * maxResolution) / height);
            height = maxResolution;
          }
        }

        const canvas = document.createElement("canvas");
        canvas.width = width;
        canvas.height = height;
        
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          return reject(new Error("Failed to get canvas 2d context"));
        }

        ctx.drawImage(img, 0, 0, width, height);
        
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              return reject(new Error("Canvas to Blob conversion failed"));
            }
            
            // Generate new filename with .webp extension
            const lastDotIndex = file.name.lastIndexOf(".");
            const baseName = lastDotIndex !== -1 ? file.name.substring(0, lastDotIndex) : file.name;
            const newName = `${baseName}.webp`;
            
            const newFile = new File([blob], newName, {
              type: "image/webp",
              lastModified: Date.now(),
            });
            
            resolve(newFile);
          },
          "image/webp",
          0.8 // 80% quality compression
        );
      };
      
      img.onerror = (error) => reject(error);
    };
    
    reader.onerror = (error) => reject(error);
  });
};
