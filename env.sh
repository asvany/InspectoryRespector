if ! [[ -z "${PY_ENV}" ]]; then
    echo "PY_ENV already set to $PY_ENV"
    return
fi


export PY_ENV=.py_env
if [ ! -f $PY_ENV/bin/python ];
then
    virtualenv -p python3 $PY_ENV || return "python not found"
fi

echo "Activating virtualenv $PY_ENV"
source $PY_ENV/bin/activate
echo "Installing requirements"
pip install -r requirements.txt
echo "Compiling protocol buffer"
protoc --go_out=. --python_out=ir_protocol_py ir_record.proto

echo "Ready to go!"

