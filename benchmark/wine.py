import lightgbm
import numpy as np
import os
import sklearn.datasets as ds

d = ds.load_wine()

X = d['data']

X = np.floor(X).astype(np.int)
params = {
    'objective': 'multiclass',
    'num_class': 3,
    'max_bin': 255,
    'num_trees': 200,
    'learning_rate': 0.1,
    'num_leaves': 10,
}

labels = d['target']
train = lightgbm.Dataset(data=X, label=d['target'], categorical_feature=list(range(X.shape[1])))
gbm = lightgbm.train(params=params, train_set=train)
test = np.vstack([X])
gbm.save_model(os.path.join('..', 'testdata', 'lgwine.model'))
pred = gbm.predict(data=test, raw_score=True)
np.savetxt(os.path.join('..', 'testdata', 'lgwine_true_predictions.txt'), pred, delimiter='\t')
ds.dump_svmlight_file(test, [0] * test.shape[0], os.path.join('..', 'testdata', 'wine_test.libsvm'))
