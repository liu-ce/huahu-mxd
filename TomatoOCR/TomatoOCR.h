#ifndef TomatoOCR_H
#define TomatoOCR_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct TomatoOCR TomatoOCR;

TomatoOCR *TomatoOCR_new();

void setMode(TomatoOCR *obj, char *mode);

void setHttpIntervalTime(TomatoOCR *obj, int second);

int init(TomatoOCR *obj, char *model_path);
         
const char* setLicense(TomatoOCR *obj, char *license, char *remark);

void setRecType(TomatoOCR *obj, char *rec_type);

void setRunMode(TomatoOCR *obj, char *mode);

void setDetBoxType(TomatoOCR *obj, char *det_box_type);

void setDetScaleRatio(TomatoOCR *obj, float det_scale_ratio);

void setDetUnclipRatio(TomatoOCR *obj, float det_unclip_ratio);

void setRecScoreThreshold(TomatoOCR *obj, float rec_score_threshold);

void setReturnType(TomatoOCR *obj, char *return_type);

void setBinaryThresh(TomatoOCR *obj, int binary_thresh);


void setFilterColor(TomatoOCR *obj, char *filterColor, char *backgroundColor);

void setBackgroundColor(TomatoOCR *obj, char *backgroundColor);

void setFilterColorPath(TomatoOCR *obj, char *filterColorPath);


const char* ocrFile(TomatoOCR* obj, char *image_path, int type);

const char* ocrImageData(TomatoOCR* obj, char *image_data, int width, int height, int type);

const char* findTapPoint(TomatoOCR* obj, char *data);

const char* findTapPoints(TomatoOCR* Obj, char *data);

void release(TomatoOCR *obj);



int initYolo(TomatoOCR *obj, char *key,
         char *yolo_secret,
         char *yolo_model_path,
         char *yolo_label_path);

const char* yoloFile(TomatoOCR* obj, char *key, char *image_path, int target_size, float score_threshold, float num_score_threshold);

const char* yoloImageData(TomatoOCR* obj, char *key, char *image_data, int width, int height, int target_size, float score_threshold, float num_score_threshold);

void releaseYolo(TomatoOCR *obj, char *key);


#ifdef __cplusplus
}
#endif

#endif // TomatoOCR_H
